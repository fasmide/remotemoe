package routertwo

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// NamedRoute implements Routable and is used when people want to create
// more human friendly hostnames for their tunnels
type NamedRoute struct {
	// Name, the FQDN
	Name string

	// Owner's pubkey fingerprint
	Owner string

	// A namedroute must know the router it was added to
	// in order to pass DialContext calls when Dialed
	router *Router
}

// NewName sets up and returns a *NamedRoute which can be added the router
func NewName(s string, r Routable) *NamedRoute {
	// ensure all names are lowercased
	s = strings.ToLower(s)

	return &NamedRoute{
		Owner: r.FQDN(),
		Name:  s,
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

	return n.router.DialContext(ctx, network, address)
}

// Replaced for NamedRoutes means deleting the NamedRoute for good and really should not
// happen - only in the case that a user tries to steal another users pubkey.hostname name -
// an when the guy with the actural key comes online - this Replaced is called which will remove it from the database
func (n *NamedRoute) Replaced() {
	panic("fix me")
}
