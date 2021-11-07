package rewrite

import (
	"github.com/fasmide/remotemoe/http"
	"github.com/fasmide/remotemoe/routertwo"
	"github.com/spf13/cobra"
)

// List displays a list of active http configurations
func List(r routertwo.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "Lists active matches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			directions := http.List(r)
			for _, d := range directions {
				cmd.Printf("\t %+v\n", d)
			}

			if len(directions) == 0 {
				cmd.Printf("No http matches added...\n")
			}

			return nil
		},
	}

	return c

}
