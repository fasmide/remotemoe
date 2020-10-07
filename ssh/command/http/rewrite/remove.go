package rewrite

import (
	"fmt"
	"net/url"

	"github.com/fasmide/remotemoe/http"
	"github.com/spf13/cobra"
)

// Remove will delete an active rewrite
func Remove() *cobra.Command {
	c := &cobra.Command{
		Use:   "remove",
		Short: "Remove active rewrites",
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := url.Parse(args[0])
			if err != nil {
				return fmt.Errorf("unable to parse url: %w", err)
			}

			d := http.Direction{}
			err = d.FromURL(u)
			if err != nil {
				return fmt.Errorf("unable to parse url: %w", err)
			}

			cmd.Printf("removing %+v", d)
			http.Remove(d)

			return nil
		},
		Args: cobra.ExactArgs(1),
	}

	return c

}
