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

// Calls a function until it succeeds, waiting an exponentially increasing amount of time between calls.
// An initial backoff of 0 means the waiting time does not increase exponentially (u