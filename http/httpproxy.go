package http

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
)

type HttpProxy struct {
	httputil.ReverseProxy
}

// Initialize sets up this proxy's transport to dial though
// Router instead of doing classic network dials
func (h *HttpProxy) Initialize() {
	transport := &http.Transport{
		DialContext:           router.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// This director will try to set r.URL to something
	// usefull based on the "virtualhost" and the destination tcp port
	h.Director = func(r *http.Request) {

		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}

		localAddr := r.Context().Value(localAddr("localaddr")).(string)
		_, dstPort, _ := net.SplitHostPort(localAddr)

		r.URL.Host = fmt.Sprintf("%s:%s", host, dstPort)

		// cant possibly fail right? :)
		dPort, _ := strconv.Atoi(dstPort)

		// services.Ports should map 80 into http, 443 into https and so on
		r.URL.Scheme = services.Ports[dPort]
	}

	h.Transport = transport

}
