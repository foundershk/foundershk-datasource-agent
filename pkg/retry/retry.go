package retry

import (
	"math"
	"time"

	"github.com/grafana/pdc-agent/pkg/random"
)

type Opts struct {
	MaxBackoff     time.Duration
	InitialBackoff time.Duration
}

// Calls a function until it succeeds, waiting an exponentially increasing amount of time betwee