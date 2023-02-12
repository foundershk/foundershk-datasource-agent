package retry

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestForever(t *testing.T) {
	t.Parallel()

	t.Run("sh