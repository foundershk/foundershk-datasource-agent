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
			descr