package http

import "github.com/spf13/cobra"

// List displays a list of active http configurations
func List() *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "Lists active matches",
		Run: func(cmd *cobra.Command, _ []string) {

		},
	}

	c.LocalFlags().BoolP("default", "d", false, "also display default matches")

	return c

}
