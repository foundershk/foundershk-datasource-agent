
package ssh

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/grafana/dskit/services"
	"github.com/grafana/pdc-agent/pkg/pdc"
	"github.com/grafana/pdc-agent/pkg/retry"
)

const (
	// The exit code sent by the pdc server when the connection limit is reached.
	ConnectionLimitReachedCode = 254
)

// Config represents all configurable properties of the ssh package.
type Config struct {
	Args []string // deprecated

	KeyFile    string
	SSHFlags   []string // Additional flags to be passed to ssh(1). e.g. --ssh-flag="-vvv" --ssh-flag="-L 80:localhost:80"
	Port       int
	LogLevel   int
	PDC        pdc.Config
	LegacyMode bool
	// ForceKeyFileOverwrite forces a new ssh key pair to be generated.
	ForceKeyFileOverwrite bool
	URL                   *url.URL
}

// DefaultConfig returns a Config with some sensible defaults set
func DefaultConfig() *Config {
	root, err := os.UserHomeDir()
	if err != nil {
		// Use relative path (should not happen)
		root = ""
	}
	return &Config{
		Port:     22,
		LogLevel: 2,
		PDC:      pdc.Config{},
		KeyFile:  path.Join(root, ".ssh/grafana_pdc"),
	}
}

func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	var deprecatedInt int

	def := DefaultConfig()

	cfg.SSHFlags = []string{}
	f.StringVar(&cfg.KeyFile, "ssh-key-file", def.KeyFile, "The path to the SSH key file.")
	f.IntVar(&deprecatedInt, "log-level", def.LogLevel, "[DEPRECATED] Use the log.level flag. The level of log verbosity. The maximum is 3.")
	// use default log level if invalid
	if cfg.LogLevel > 3 {
		cfg.LogLevel = def.LogLevel
	}
	f.Func("ssh-flag", "Additional flags to be passed to ssh. Can be set more than once.", cfg.addSSHFlag)
	f.BoolVar(&cfg.ForceKeyFileOverwrite, "force-key-file-overwrite", false, "Force a new ssh key pair to be generated")
}

func (cfg Config) KeyFileDir() string {
	dir, _ := path.Split(cfg.KeyFile)
	return dir
}

func (cfg *Config) addSSHFlag(s string) error {
	cfg.SSHFlags = append(cfg.SSHFlags, s)
	return nil
}

// Client is a client for ssh. It configures and runs ssh commands
type Client struct {
	*services.BasicService
	cfg    *Config
	SSHCmd string // SSH command to run, defaults to "ssh". Require for testing.
	logger log.Logger
	km     *KeyManager
}

// NewClient returns a new SSH client in an idle state
func NewClient(cfg *Config, logger log.Logger, km *KeyManager) *Client {
	client := &Client{
		cfg:    cfg,
		SSHCmd: "ssh",
		logger: logger,
		km:     km,
	}

	client.BasicService = services.NewIdleService(client.starting, client.stopping)
	return client
}

func (s *Client) starting(ctx context.Context) error {
	level.Info(s.logger).Log("msg", "starting ssh client")

	// check keys and cert validity before start, create new cert if required
	// This will exit if it fails, rather than endlessly retrying to sign keys.
	if s.km != nil {
		err := s.km.CreateKeys(ctx)
		if err != nil {
			level.Error(s.logger).Log("msg", "could not check or generate certificate", "error", err)
			return err
		}
	}

	// Attempt to parse SSH flags before triggering the goroutine, so we can exit
	// if the parsing fails
	flags, err := s.SSHFlagsFromConfig()
	if err != nil {
		level.Error(s.logger).Log("msg", fmt.Sprintf("could not parse flags: %s", err))
		return err
	}
	level.Debug(s.logger).Log("msg", fmt.Sprintf("parsed flags: %s", flags))

	retryOpts := retry.Opts{MaxBackoff: 16 * time.Second, InitialBackoff: 1 * time.Second}
	go retry.Forever(retryOpts, func() error {
		cmd := exec.CommandContext(ctx, s.SSHCmd, flags...)
		loggerWriter := newLoggerWriterAdapter(s.logger)
		cmd.Stdout = loggerWriter
		cmd.Stderr = loggerWriter
		_ = cmd.Run()
		if ctx.Err() != nil {
			return nil // context was canceled
		}

		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == ConnectionLimitReachedCode {
			level.Info(s.logger).Log("msg", "limit of connections for stack and network reached. exiting")
			os.Exit(1)
		}

		level.Error(s.logger).Log("msg", "ssh client exited. restarting")

		// Check keys and cert validity before restart, create new cert if required.
		// This covers the case where a certificate has become invalid since the last start.
		// Do not return here: we want to keep trying to connect in case the PDC API
		// is temporarily unavailable.
		if s.km != nil {