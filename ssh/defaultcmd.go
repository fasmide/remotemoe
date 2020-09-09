package ssh

import (
	"github.com/fasmide/remotemoe/ssh/command"
	"github.com/spf13/cobra"
)

// DefaultCmd is the default top level command, embedding all others
func DefaultCmd(session *Session) *cobra.Command {
	c := &cobra.Command{
		Use:          "main",
		SilenceUsage: true,
	}

	c.AddCommand(command.Firsttime())
	c.AddCommand(command.Close(session))
	c.AddCommand(command.Session(session))
	c.AddCommand(command.Host(session))
	c.AddCommand(command.Access(session))
	return c
}
