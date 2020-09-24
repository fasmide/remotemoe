package command

import (
	"fmt"

	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/ssh/command/host"
	"github.com/spf13/cobra"
)

// Host returns a *cobra.Command that enables the user to mange custom hosts
func Host(r router.Routable) *cobra.Command {
	top := &cobra.Command{
		Use:   "host",
		Short: "Manage hostnames",
	}

	top.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List hostnames",
		RunE: func(cmd *cobra.Command, _ []string) error {
			namedRoutes, err := router.Names(r)
			if err != nil {
				return fmt.Errorf("unable to lookup your custom names: %w", err)
			}

			cmd.Printf("Active hostnames:\n")

			cmd.Printf("%s (fixed)\n", r.FQDN())
			for _, nr := range namedRoutes {
				cmd.Printf("%s\n", nr.FQDN())
			}

			return nil
		},
	})

	top.AddCommand(host.Remove(r))
	top.AddCommand(host.Add(r))

	return top
}
