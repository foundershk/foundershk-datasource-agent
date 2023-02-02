package random

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestRange(t *testing.T) {
	t.Parallel()

	t.Run("sanity checks", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, 0, Range(0, 0))
		assert.Equal(t, 1, Range(1, 1))

		require.Eventually(t, func() bool {
			return Range(0, 1) == 0
		}, 500*time.Millisecond, 50*time.Microsecond)

		require.Eventually(t, func() 