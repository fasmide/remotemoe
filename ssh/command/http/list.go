package http

import (
	"fmt"

	"github.com/fasmide/remotemoe/router"
	"github.com/spf13/cobra"
)

// List displays a list of active http configurations
func List(r router.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "Lists active matches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			hosts, err := router.Names(r)
			if err != nil {
				return fmt.Errorf("could not lookup hosts: %w", err)
			}
		},
	}

	c.LocalFlags().BoolP("default", "d", false, "also display default matches")

	return c

}
