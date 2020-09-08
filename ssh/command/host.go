package command

import (
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/ssh/command/host"
	"github.com/spf13/cobra"
)

func Host(r router.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   "host",
		Short: "Manage hostnames",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	c.AddCommand(host.Remove(r))

	return c
}
