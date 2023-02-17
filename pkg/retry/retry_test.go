package retry

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestForever(t *testing.T) {
	t.Parallel()

	t.Run("should retry until the function succeeds", func(t *testing.T) {
		t.Parallel()

		attempts := 0

		retryOpts := Opts{MaxBackoff: 100 * time.Second, InitialBackoff: 0 * time.Second}
		Forever(retryOpts, func() error {
			a