
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/grafana/dskit/services"
	"github.com/grafana/pdc-agent/pkg/pdc"
	"github.com/grafana/pdc-agent/pkg/ssh"
)

// Values set by goreleaser during the build process using ldflags.
// https://goreleaser.com/cookbooks/using-main.version/
var (
	// Current Git tag (the v prefix is stripped) or the name of the snapshot, if you're using the --snapshot flag
	version string
	// Current git commit SHA
	commit string
	// Date in the RFC3339 format
	date string
)

const logLevelinfo = "info"

type mainFlags struct {
	PrintHelp bool
	LogLevel  string
	Cluster   string
	Domain    string

	// The fields below were added to make local development easier.
	//
	// DevMode is true when the agent is being run locally while someone is working on it.
	DevMode bool
}

func (mf *mainFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&mf.PrintHelp, "h", false, "Print help")
	fs.StringVar(&mf.LogLevel, "log.level", logLevelinfo, `"debug", "info", "warn" or "error"`)
	fs.StringVar(&mf.Cluster, "cluster", "", "the PDC cluster to connect to use")
	fs.StringVar(&mf.Domain, "domain", "grafana.net", "the domain of the PDC cluster")
	fs.BoolVar(&mf.DevMode, "dev-mode", false, "[DEVELOPMENT ONLY] run the agent in development mode")
}

func logLevelToSSHLogLevel(level string) (int, error) {
	switch level {
	case "error", "warn", "info":
		return 0, nil
	case "debug":
		return 3, nil
	default:
		return -1, fmt.Errorf("invalid log level: %s", level)
	}
}

// Tries to get the openssh version. Returns "UNKNOWN" on error.
func tryGetOpenSSHVersion() string {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	buffer := bytes.NewBuffer([]byte{})

	cmd := exec.CommandContext(timeoutCtx, "ssh", "-V")
	// ssh -V outputs to stderr.
	cmd.Stderr = buffer

	if err := cmd.Run(); err != nil {
		return "UNKNOWN"
	}

	// ssh -V adds \n to the end of the output.
	return strings.Replace(buffer.String(), "\n", "", 1)
}

func main() {
	sshConfig := ssh.DefaultConfig()
	mf := &mainFlags{}
	pdcClientCfg := &pdc.Config{}

	usageFn, err := parseFlags(mf.RegisterFlags, sshConfig.RegisterFlags, pdcClientCfg.RegisterFlags)
	if err != nil {
		fmt.Println("cannot parse flags")
		os.Exit(1)
	}

	sshConfig.Args = os.Args[1:]
	sshConfig.LogLevel, err = logLevelToSSHLogLevel(mf.LogLevel)
	if err != nil {
		usageFn()
		fmt.Printf("setting log level: %s\n", err)
		os.Exit(1)
	}

	logger := setupLogger(mf.LogLevel)

	level.Info(logger).Log("msg", "PDC agent info",
		"version", fmt.Sprintf("v%s", version),
		"commit", commit,
		"date", date,
		"ssh version", tryGetOpenSSHVersion(),
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
	)

	if mf.PrintHelp {
		usageFn()
		return
	}

	if inLegacyMode() {
		sshConfig.LegacyMode = true
		err = runLegacyMode(sshConfig)
		if err != nil {
			fmt.Printf("error: %s", err)
			os.Exit(1)
		}
		return
	}

	apiURL, gatewayURL, err := createURLsFromCluster(mf.Cluster, mf.Domain)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	pdcClientCfg.URL = apiURL
	sshConfig.PDC = *pdcClientCfg
	sshConfig.URL = gatewayURL

	if mf.DevMode {
		setDevelopmentConfig(sshConfig, pdcClientCfg)
	}

	err = run(logger, sshConfig, pdcClientCfg)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

}

// Configures the agent for local development
func setDevelopmentConfig(sshCfg *ssh.Config, pdcClientCfg *pdc.Config) {
	pdcClientCfg.URL, _ = url.Parse("http://localhost:9181")

	pdcClientCfg.DevHeaders = map[string]string{
		"X-Scope-OrgID":      pdcClientCfg.HostedGrafanaID,
		"X-Access-Policy-ID": pdcClientCfg.DevNetwork,
	}
	pdcClientCfg.SignPublicKeyEndpoint = "/api/v1/sign-public-key"

	sshCfg.Port = 2244
	sshCfg.URL, _ = url.Parse("localhost")
	sshCfg.PDC = *pdcClientCfg
}

func run(logger log.Logger, sshConfig *ssh.Config, pdcConfig *pdc.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pdcClient, err := pdc.NewClient(pdcConfig, logger)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot initialise PDC client: %s", err))
		return err
	}

	km := ssh.NewKeyManager(sshConfig, logger, pdcClient)

	// Create the SSH Service. KeyManager must be in running state when passed to ssh.NewClient
	sshClient := ssh.NewClient(sshConfig, logger, km)
	// Start the ssh client
	err = services.StartAndAwaitRunning(ctx, sshClient)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot start ssh client: %s", err))
		return err
	}

	// Wait for the ssh client to exit
	_ = sshClient.AwaitTerminated(context.Background())

	return nil
}

func createURLsFromCluster(cluster string, domain string) (api *url.URL, gateway *url.URL, err error) {