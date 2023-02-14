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

		retryOp