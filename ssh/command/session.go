package command

import "github.com/spf13/cobra"

type Data interface {
	Forwards() map[uint32]struct{}
}

func Session(d Data) *cobra.Command {
	c := &cobra.Command{
		Use:   "session",
		Short: "Info about this session",
	}

	c.AddCommand(Forwards(d))
	c.AddCommand(Autossh(d))
	c.AddCommand(Unitfile(d))

	return c
}
