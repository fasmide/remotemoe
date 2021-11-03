package command

import (
	"fmt"

	"github.com/fasmide/remotemoe/routertwo"
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

// Whoami tells the user their FQDN
func Whoami(r routertwo.Routable) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: fmt.Sprintf("Prints this connection <hash>.%s hostname", services.Hostname),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("%s\n", r.FQDN())
		},
	}
}
