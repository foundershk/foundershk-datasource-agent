package ssh

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/grafana/pdc-agent/pkg/pdc"
	"github.com/mikesmitty/edkey"
	"golang.org/x/crypto/ssh"
)

const (
	// SSHKeySize is the size of the SSH key.
	SSHKeySize     = 4096
	KnownHostsFile = "grafana_pdc_known_hosts"
)

// TODO
// KeyManager implements KeyManager. If needed, it gets new certificates signed
// by the PDC API.
//
// If the service starts successfully, then the key and cert files will exist.
// It will attempt to reuse existing keys and certs if they exist.
type KeyManager struct {
	cfg    *Config
	client pdc.Client
	logger log.Logger
}

// NewKeyManager returns a new KeyManager in an idle state
func NewKeyManager(cfg *Config, logger log.Logger, client pdc.Client) *KeyManager {
	km := KeyManager{
		cfg:    cfg,
		client: client,
		logger: logger,
	}

	return &km
}

func (km *KeyManager) CreateKeys(ctx context.Context) error {
	level.Info(km.logger).Log("msg", "starting key manager")

	newCertRequired, err := km.ensureKeysExist(km.cfg.ForceKeyFileOverwrite)
	if err != nil {
		return err
	}

	argumentHash := km.argumentsHash()
	if km.argumentsHashIsDifferent(argumentHash) {
		level.Info(km.logger).Log("msg", fmt.Sprintf("fetching new certificate: agent arguments changed hash=%s", argumentHash))
		newCertRequired = true
	}

	if err := km.ensureCertExists(ctx, newCertRequired); err != nil {
		return fmt.Errorf("ensuring certificate exists: %w", err)
	}

	if err := km.writeHashFile([]byte(argumentHash)); err != nil {
		return fmt.Errorf("writing to hash file: %w", err)
	}

	return nil
}

// EnsureCertExists checks for the existence of a valid SSH certificate and
// regenerates one if it cannot find one, or if forceCreate is true.
func (km KeyManager) ensureCertExists(ctx context.Context, forceCreate bool) error {
	newCertRequired := forceCreate

	if newCertRequired {
		err := km.generateCert(ctx)
		if err