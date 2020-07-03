package router

import (
	"context"
	"fmt"
	"net"
	"time"
)

type namedRoute struct {
	// Owner's pubkey fingerprint
	Owner string `storm:"index"`

	// Name: any fqdn domain for this route
	Name string `storm:"unique"`

	// LastSeen is used when garbage collecting
	LastSeen time.Time

	Created time.Time
}

func (n *namedRoute) NewName(s string, r Routable) *namedRoute {
	return &namedRoute{
		Owner:    r.FQDN(),
		Name:     s,
		LastSeen: time.Now(),
		Created:  time.Now(),
	}
}

func (n *namedRoute) FQDN() string {
	return n.Owner
}

func (n *namedRoute) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	_, p, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("namedRoute: cannot split host from port on '%s': %w", address, err)
	}

	address = net.JoinHostPort(n.Owner, p)

	return DialContext(ctx, network, address)
}

// Replaced fulfills the Routable interface but does not make much
// logical sense to a named route as users should not be able to takeover existing
// named routes
func (n *namedRoute) Replaced() {
	// does this indicate someone was trying to take over an endpoint ?
}
