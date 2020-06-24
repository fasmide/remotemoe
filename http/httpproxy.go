package http

import (
	"fmt"
	"net"
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

	// This directory will try to set r.URL to something
	// usefull based on the "virtualhost" and the destination tcp port
	h.Director = func(r *http.Request) {
		// maybe map the portnumer into a scheme ?
		// we are fixed to having tls on 443 and non-tls on 80 anyway
		r.URL.Scheme = "http"

		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}

		localAddr := r.Context().Value(localAddr("localaddr")).(string)
		_, dstPort, _ := net.SplitHostPort(localAddr)

		r.URL.Host = fmt.Sprintf("%s:%s", host, dstPort)
	}

	h.Transport = transport

}
