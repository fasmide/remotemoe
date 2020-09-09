package command

import "github.com/spf13/cobra"

// Forwarding are things that forwards ports
type Forwarding interface {
	Forwards() map[uint32]struct{}
}

// Session returns a command that tells the user information about the current session
func Session(d Forwarding) *cobra.Command {
	c := &cobra.Command{
		Use:   "session",
		Short: "Info about this session",
	}

	c.AddCommand(Forwards(d))
	c.AddCommand(Autossh(d))
	c.AddCommand(Unitfile(d))

	return c
}
