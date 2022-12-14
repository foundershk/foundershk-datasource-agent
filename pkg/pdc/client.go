
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