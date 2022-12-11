package httpclient

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// UserAgentTransport provides a transport with a set user-agent. It wraps
// http.DefaultTran