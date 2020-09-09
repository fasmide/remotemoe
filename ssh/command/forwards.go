package command

import (
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

func Forwards(d Data) *cobra.Command {
	return &cobra.Command{
		Use:   "forwards",
		Short: "Lists active forwarded ports",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Forwarded ports: (port/protocol)")

			f := d.Forwards()
			for p := range f {
				s, exists := services.Ports[int(p)]
				if !exists {
					s = "unknown"
				}

				cmd.Printf("%d/%s\n", p, s)
			}

			if len(f) == 0 {
				cmd.Printf("None\n")
			}
		},
	}
}
