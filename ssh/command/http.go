package command

import (
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/ssh/command/http"
	"github.com/spf13/cobra"
)

const longHelp = `Configure the behavior of the front-facing HTTP proxy.

By default, the HTTP proxy tries pass requests on by doing the least amount
of changes possible, it does not change the Host header, TCP port, or strip SSL.

For example, if an https request is accepted, remotemoe will
try to make a TLS connection upstream inside the ssh session.

Requests to non-default HTTP ports, e.g., http://xyz.domain.tld:8080 will be upstreamed
to the same non-default port 8080 or fail if this port is not available in
the ssh session.

An X-Forwarded-For header is added, which should be the only default change.

Directions can be changed, for instance, you could drop ssl inside the ssh tunnel,
by adding a direction from https://somehost.tld/ to http://somehost.tld:8080/.

The HTTP proxy only dials inside remotemoe and cannot upstream HTTP requests outside tunnels.

`

// HTTP is the toplevel command user management of the http proxy
func HTTP(session router.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   "http",
		Short: "HTTP proxy management",
		Long:  longHelp,
	}

	c.AddCommand(http.Rewrite(session))

	return c
}
