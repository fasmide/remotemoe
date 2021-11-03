package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

// Proxy reverse proxies requests though router
type Proxy struct {
	httputil.ReverseProxy
}

// Dialer interface describes the minimun methods a Proxy needs
type Dialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

// Initialize sets up this proxy's transport to dial though
// Router instead of doing classic network dials
func (h *Proxy) Initialize(router Dialer) {
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

	h.Director = director
	h.Transport = transport

}
