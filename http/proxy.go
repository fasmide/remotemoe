package http

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/fasmide/remotemoe/router"
)

// Proxy reverse proxies requests though router
type Proxy struct {
	httputil.ReverseProxy
}

// Initialize sets up this proxy's transport to dial though
// Router instead of doing classic network dials
func (h *Proxy) Initialize() {
	transport := &http.Transport{
		DialContext:           router.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxConnsPerHost:       10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// TLS inside the ssh tunnel will not be able to provide any valid certificate so ..
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// This director will try to set r.URL to something
	// useful based on the "virtualhost" and the destination tcp port
	h.Director = director

	h.Transport = transport

}
