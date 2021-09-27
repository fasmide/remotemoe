package command

import (
	"github.com/fasmide/remotemoe/buildvars"
	"github.com/fasmide/remotemoe/buildvars/israce"
	"github.com/spf13/cobra"
)

// Version reports buildvars and maybe a future version number
func Version() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Version reports buildvars",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Git Repository:   %s\n", buildvars.GitRepository)
			cmd.Printf("Git Branch:       %s\n", buildvars.GitBranch)
			cmd.Println()
			cmd.Printf("Git Commit:       %s\n", buildvars.GitCommit)
			cmd.Printf("Git Commit Date:  %s\n", buildvars.GitCommitDate)
			cmd.Printf("Git Link:         https://github.com/fasmide/remotemoe/commit/%s\n", buildvars.GitCommit)
			cmd.Println()
			cmd.Print("Git Local Changes: ")
			if len(buildvars.GitPorcelain) == 0 {
				cmd.Print("false\n")
			} else {
				cmd.Printf("true\n%s\n", buildvars.GitPorcelain)
			}
			cmd.Println()
			cmd.Printf("Race Detection:   %t\n", israce.Enabled)
		},
	}
}
