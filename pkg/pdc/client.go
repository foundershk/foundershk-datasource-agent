
package pdc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/pdc-agent/pkg/httpclient"
	"github.com/hashicorp/go-retryablehttp"

	"golang.org/x/crypto/ssh"
)

var (
	// ErrInternal indicates the item could not be processed.
	ErrInternal = errors.New("internal error")
	// ErrInvalidCredentials indicates the auth token is incorrect
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Config describes all properties that can be configured for the PDC package
type Config struct {
	Token           string
	HostedGrafanaID string
	URL             *url.URL
	RetryMax        int

	// The PDC api endpoint used to sign public keys.
	// It is not a constant only to make it easier to override the endpoint in local development.
	SignPublicKeyEndpoint string

	// Used for local development.
	// Contains headers that are included in each http request send to the pdc api.
	DevHeaders map[string]string

	// Used for local development.
	// DevNetwork is the network that the agent will connect to.
	DevNetwork string
}

func (cfg *Config) RegisterFlags(fs *flag.FlagSet) {
	var deprecated string
	fs.StringVar(&cfg.Token, "token", "", "The token to use to authenticate with Grafana Cloud. It must have the pdc-signing:write scope")
	fs.StringVar(&cfg.HostedGrafanaID, "gcloud-hosted-grafana-id", "", "The ID of the Hosted Grafana instance to connect to")
	fs.StringVar(&cfg.DevNetwork, "dev-network", "", "[DEVELOPMENT ONLY] the network the agent will connect to")
	fs.StringVar(&deprecated, "network", "", "DEPRECATED: The name of the PDC network to connect to")
}

// Client is a PDC API client
type Client interface {
	SignSSHKey(ctx context.Context, key []byte) (*SigningResponse, error)
}

// SigningResponse is the response received from a SSH key signing request
type SigningResponse struct {
	Certificate ssh.Certificate
	KnownHosts  []byte
}

func (sr *SigningResponse) UnmarshalJSON(data []byte) error {
	target := struct {
		Certificate string `json:"certificate"`
		KnownHosts  string `json:"known_hosts"`
	}{}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	err := dec.Decode(&target)
	if err != nil {
		return err
	}

	block, rest := pem.Decode([]byte(target.Certificate))
	if block == nil {
		return fmt.Errorf("failed to pem-decode certificate: %w", err)
	}
	if len(rest) > 0 {
		return fmt.Errorf("only expected one PEM")
	}
	pk, _, _, _, err := ssh.ParseAuthorizedKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	cert, ok := pk.(*ssh.Certificate)
	if !ok {
		return errors.New("public key is not an SSH certificate")
	}

	sr.KnownHosts = []byte(target.KnownHosts)
	sr.Certificate = *cert
	return nil
}

// NewClient returns a new Client
func NewClient(cfg *Config, logger log.Logger) (Client, error) {
	if cfg.URL == nil {
		return nil, errors.New("-api-url cannot be nil")
	}

	// If the value has not been set for testing.
	if cfg.SignPublicKeyEndpoint == "" {
		cfg.SignPublicKeyEndpoint = "/pdc/api/v1/sign-public-key"
	}

	rc := retryablehttp.NewClient()
	if cfg.RetryMax != 0 {
		rc.RetryMax = cfg.RetryMax
	}
	rc.Logger = &logAdapter{logger}
	rc.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy
	hc := rc.StandardClient()

	hc.Transport = httpclient.UserAgentTransport(hc.Transport)

	return &pdcClient{
		cfg:        cfg,
		httpClient: hc,
		logger:     logger,
	}, nil
}

type pdcClient struct {
	cfg        *Config
	httpClient *http.Client
	logger     log.Logger
}

func (c *pdcClient) SignSSHKey(ctx context.Context, key []byte) (*SigningResponse, error) {
	resp, err := c.call(ctx, http.MethodPost, c.cfg.SignPublicKeyEndpoint, nil, map[string]string{
		"publicKey": string(key),
	})
	if err != nil {
		return nil, err
	}

	sr := &SigningResponse{}
	err = sr.UnmarshalJSON(resp)
	if err != nil {
		return nil, err
	}