package host

import (
	"fmt"

	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

func Add(r router.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   fmt.Sprintf("add host.%s [host2.domain.tld] ...", services.Hostname),
		Short: "Add hostname(s)",
		Args:  cobra.MinimumNArgs(1),
		Long:  "Add hostname(s)\n\nAdd as many hostnames as needed.\nBring your own domains by setting up DNS records appropriately.",
		Run: func(cmd *cobra.Command, args []string) {
			for _, n := range args {
				namedRoute := router.NewName(n, r)

				err := router.Add(namedRoute)
				if err != nil {
					cmd.Printf("%s could not be added: %s\n", n, err)
					continue
				}

				cmd.Printf("%s is active.\n", namedRoute.FQDN())
			}
		},
	}

	return c
}
