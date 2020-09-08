package command

import "github.com/spf13/cobra"

func Coffie() *cobra.Command {
	return &cobra.Command{
		Use:   "coffie",
		Short: "have some lovely coffie",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("here you go")
		},
	}
}
