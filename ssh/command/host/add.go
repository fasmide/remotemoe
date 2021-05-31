package host

import (
	"fmt"

	"github.com/fasmide/remotemoe/routertwo"
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

// Add returns a cobra.Command which can add custom hostnames
func Add(r routertwo.Routable, router *routertwo.Router) *cobra.Command {
	c := &cobra.Command{
		Use:   fmt.Sprintf("add host.%s [host2.domain.tld] ...", services.Hostname),
		Short: "Add hostname(s)",
		Args:  cobra.MinimumNArgs(1),
		Long:  "Add hostname(s)\n\nAdd as many hostnames as needed.\nBring your own domains by setting up DNS records appropriately.",
		Run: func(cmd *cobra.Command, args []string) {
			for _, n := range args {
				namedRoute := routertwo.NewName(n, r)

				err := router.AddName(namedRoute)
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
