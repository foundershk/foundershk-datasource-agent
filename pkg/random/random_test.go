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

	t.Run("sanity checks"