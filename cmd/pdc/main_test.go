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
		level         