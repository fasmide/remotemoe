package rewrite

import (
	"fmt"

	"github.com/fasmide/remotemoe/http"
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

			// initialize the hosts we would like to search for
			// begining with this sessions FQDN and then add other namedroutes
			searchHosts := []string{r.FQDN()}
			for _, namedRoute := range hosts {
				searchHosts = append(searchHosts, namedRoute.FQDN())
			}

			directions := http.List(searchHosts...)
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
