package httpclient

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// UserAgentTransport provides a transport with a set user-agent. It wraps
// http.DefaultTransport if rt is nil
func UserAgentTransport(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	ua := "pdc-httpc