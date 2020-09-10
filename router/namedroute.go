package router

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// NamedRoute implements Routable and is used when people want to create
// more human friendly hostnames for their tunnels
type NamedRoute struct {
	// Owner's pubkey fingerprint
	Owner string `storm:"index"`

	// Name: any fqdn domain for this route
	Name string `storm:"id"`

	// LastSeen is used when garbage collecting
	LastSeen time.Time

	Created time.Time
}

// NewName sets up and returns a *NamedRoute which can be added the router
func NewName(s string, r Routable) *NamedRoute {
	// ensure all names are lowercased
	s = strings.ToLower(s)

	return &NamedRoute{
		Owner:    r.FQDN(),
		Name:     s,
		LastSeen: time.Now(),
		Created:  time.Now(),
	}
}

// FQDN returns the fully qualified domain name for this route
func (n *NamedRoute) FQDN() string {
	return n.Name
}

// DialContext on a NamedRoute looks up the correct tunnel in the router and uses its DialContext
func (n *NamedRoute) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	_, p, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("NamedRoute: cannot split host from port on '%s': %w", address, err)
	}

	address = net.JoinHostPort(n.Owner, p)

	return DialContext(ctx, network, address)
}

// Replaced for NamedRoutes means deleting the NamedRoute for good and really should not
// happen - only in the case that a user tries to steal another users pubkey.hostname name -
// an when the guy with the actural key comes online - this Replaced is called which will remove it from the database
func (n *NamedRoute) Replaced() {
	err := db.DeleteStruct(n)
	if err != nil {
		log.Printf("router.*NamedRoute.Replaced(): %s could be removed: %s", n.FQDN(), err)
	}
}
