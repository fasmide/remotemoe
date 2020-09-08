package ssh

import (
	"github.com/fasmide/remotemoe/ssh/command"
	"github.com/spf13/cobra"
)

func DefaultCmd(session *Session) *cobra.Command {
	c := &cobra.Command{
		Use:          "main",
		SilenceUsage: true,
	}

	c.AddCommand(command.Coffie())
	c.AddCommand(command.Close(session))
	c.AddCommand(command.Host(session))

	return c
}
