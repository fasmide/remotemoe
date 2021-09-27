package command

import (
	"github.com/fasmide/remotemoe/buildvars"
	"github.com/spf13/cobra"
)

// Version reports buildvars and maybe a future version number
func Version() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Version reports buildvars",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Git Branch:     %s\n", buildvars.GitBranch)
			cmd.Printf("Git Commit:     %s\n", buildvars.GitHash)
			cmd.Printf("Git Date:       %s\n", buildvars.GitDate)
			cmd.Printf("Git Repository: %s\n", buildvars.GitRepository)
		},
	}
}
