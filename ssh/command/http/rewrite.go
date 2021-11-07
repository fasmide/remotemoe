package http

import (
	"github.com/fasmide/remotemoe/routertwo"
	"github.com/fasmide/remotemoe/ssh/command/http/rewrite"
	"github.com/spf13/cobra"
)

const longHelp = `
`

// Rewrite helps the user rewrite http requests as they flow though
func Rewrite(session routertwo.Routable, router *routertwo.Router) *cobra.Command {
	c := &cobra.Command{
		Use:   "rewrite",
		Short: "Rewrite scheme and port upstream",
		Long:  longHelp,
	}

	c.AddCommand(rewrite.Add(session, router))
	c.AddCommand(rewrite.List(session))
	c.AddCommand(rewrite.Remove(session))

	return c
}
