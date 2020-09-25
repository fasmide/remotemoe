package http

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/fasmide/remotemoe/services"
)

type Direction struct {
	Scheme string
	Host   string
	Port   string
}

func (d *Direction) String() string {
	return d.Scheme + "://" + d.Host + ":" + d.Port
}

func (d *Direction) FromURL(u *url.URL) error {
	// if no port specified, insert default scheme port
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	if !strings.Contains(u.Scheme, "http") {
		return fmt.Errorf("unknown scheme '%s': only http or https supported", u.Scheme)
	}

	d.Scheme = u.Scheme
	d.Host = host
	d.Port = port

	return nil
}

var matches map[string]Direction
var lock sync.RWMutex

func init() {
	matches = make(map[string]Direction)
}

func Add(m Direction, d Direction) error {
	lock.Lock()
	defer lock.Unlock()

	_, exists := matches[m.String()]
	if exists {
		return fmt.Errorf("%s already exists", m)
	}

	matches[m.String()] = d

	return nil
}

func Remove(d Direction) {
	lock.Lock()
	defer lock.Unlock()

	delete(matches, d.String())

}

func director(r *http.Request) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	localAddr := r.Context().Value(localAddr("localaddr")).(string)
	_, dstPort, _ := net.SplitHostPort(localAddr)

	r.URL.Host = fmt.Sprintf("%s:%s", host, dstPort)

	dPort, _ := strconv.Atoi(dstPort)

	// services.Ports should map 80 into http, 443 into https and so on
	r.URL.Scheme = services.Ports[dPort]

	// rewrite direction if a Match exists
	lock.RLock()
	defer lock.RUnlock()
	if dest, exists := matches[r.URL.Scheme+"://"+r.URL.Host]; exists {
		// as we are only matching scheme + "host:port"
		// we cannot just replace the URL
		// - if we did, the url would loose its /path and potential query parameters
		r.URL.Scheme = dest.Scheme
		r.URL.Host = dest.Host + ":" + dest.Port
	}
}
