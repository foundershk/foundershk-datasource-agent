package pdc_test

import (
	"encoding/json"
	"testing"

	"github.com/grafana/pdc-agent/pkg/pdc"
	"github.com/stretchr/testify/assert"
)

var cert = `
-----BEGIN CERTIFICATE-----
c3NoLWVkMjU1MTktY2VydC12MDFAb3BlbnNzaC5jb20gQUFBQUlITnp