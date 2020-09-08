package host

import (
	"fmt"

	"github.com/fasmide/remotemoe/router"
	"github.com/spf13/cobra"
)

func Remove(r router.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   "remove host.domain.tld [host2.domain.tld] ... ",
		Short: "remove hostname(s)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// try to remove all provided hosts
			for _, name := range args {
				err := router.RemoveName(name, r)
				if err != nil {
					return fmt.Errorf("could not remove %s: %s", name, err)
				}

				cmd.Printf("%s removed.\n", name)
			}

			return nil
		},
	}

	c.AddCommand(RemoveAll(r))

	return c
}
