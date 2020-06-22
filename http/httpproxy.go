package http

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/fasmide/remotemoe/router"
)

type HttpProxy struct {
	httputil.ReverseProxy

	Router *router.Router
}

// Initialize sets up this proxy's transport to dial though
// Router instead of doing classic network dials
func (h *HttpProxy) Initialize() {
	transport := &http.Transport{
		DialContext:           h.Router.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	h.Director = func(r *http.Request) {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
	}
	h.Transport = transport

}
