package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevelToSSHLogLevel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		description   string
		level         string
		expectedLevel int
		expectedErr   error
	}{
		{
			description:   "error becomes 0",
			level:         "error",
			expectedLevel: 0,
		},
		{
			description:   "warn becomes 0",
			level:         "warn",
			expectedLevel: 0,
		},
		{
			description:   "info becomes 0",
			level:         "info",
			expectedLevel: 0,
		},
		{
			description:   "debug becomes 3",
			level:         "debug",
			expectedLevel: 3,
		},
		{
			description: "unknown level, should return error",
			level:       "unknown",
			expecte