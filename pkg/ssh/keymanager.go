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
	// SSHKey