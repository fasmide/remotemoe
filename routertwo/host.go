package routertwo

import (
	"context"
	"errors"
	"net"
	"time"
)

// ErrOffline used to indicate an offline host
var ErrOffline = errors.New("peer offline")

// Host represents a host that can be offline
type Host struct {
	Routable `json:"-"`
	Name     string `json:"name"`

	// LastSeen is used when garbage collecting
	LastSeen time.Time `json:"lastseen"`
	Created  time.Time `json:"created"`
}

// FQDN returns the fully qualified domain name for this host
func (h *Host) FQDN() string {
	return h.Name
}

// DialContext dials this host
func (h *Host) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if h.Routable == nil {
		return nil, ErrOffline
	}

	return h.Routable.DialContext(ctx, network, address)
}

// Replaced indicates to this host, that it have been replaced
func (h *Host) Replaced() {
	if h.Routable != nil {
		h.Routable.Replaced()
	}
}
