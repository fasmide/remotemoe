package command

import (
	"github.com/spf13/cobra"
)

const longHelp = `Configure the behavior of the front-facing HTTP proxy.

By default, the HTTP proxy tries pass requests on by doing the least amount
of changes possible, it does not change the Host header, TCP port, or strip SSL.

For example, if an https request is accepted, 
it will try to make a TLS connection upstream inside the ssh session.

Requests to non-default HTTP ports, e.g., http://xyz.domain.tld:8080 will be upstreamed
to the same non-default port 8080 or fail if this port is not available in
the ssh session.

An X-Forwarded-For header is added, which should be the only default change.

`

func HTTP() *cobra.Command {
	c := &cobra.Command{
		Use:   "http",
		Short: "HTTP proxy management",
		Long:  longHelp,
	}

	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List hostnames",
		Run: func(cmd *cobra.Command, _ []string) {
		},
	})
	return c
}
