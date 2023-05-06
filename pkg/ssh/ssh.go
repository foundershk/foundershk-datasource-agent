
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