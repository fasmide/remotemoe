package command

import "github.com/spf13/cobra"

type Data interface {
	Forwards() map[uint32]struct{}
}

func Session(d Data) *cobra.Command {
	c := &cobra.Command{
		Use:   "session",
		Short: "Information about this session",
	}

	c.AddCommand(Forwards(d))
	return c
}
