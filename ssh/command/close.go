package command

import (
	"io"
	"log"

	"github.com/spf13/cobra"
)

func Close(c io.Closer) *cobra.Command {
	return &cobra.Command{
		Use:     "quit",
		Aliases: []string{"exit", "close"},
		Short:   "Closes the session",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("kthxbye")

			err := c.Close()
			if err != nil {
				log.Printf("ssh/command/close: %s", err)
			}
		},
	}
}
