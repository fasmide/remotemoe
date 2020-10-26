package http

import (
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/ssh/command/http/rewrite"
	"github.com/spf13/cobra"
)

const longHelp = `
`

// Rewrite helps the user rewrite http requests as they flow though
func Rewrite(session router.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   "rewrite",
		Short: "Rewrite scheme and port upstream",
		Long:  longHelp,
	}

	c.AddCommand(rewrite.Add(session))
	c.AddCommand(rewrite.List(session))
	c.AddCommand(rewrite.Remove())

	return c
}
