package command

import (
	"runtime"

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
			cmd.Printf("Go version:        %s\n", runtime.Version())
			cmd.Printf("Go arch:           %s/%s\n", runtime.GOOS, runtime.GOARCH)
			cmd.Printf("Race detection:    %t\n", israce.Enabled)

			// If buildvars have not been initialized by build.sh
			// - ship the rest
			if buildvars.Initialized != "true" {
				return
			}

			cmd.Println()
			cmd.Printf("Git Repository:    %s\n", buildvars.GitRepository)
			cmd.Printf("Git Branch:        %s\n", buildvars.GitBranch)
			cmd.Printf("Git Commit Date:   %s\n", buildvars.GitCommitDate)
			cmd.Printf("Git Commit:        %s\n", buildvars.GitCommit)
			cmd.Println()

			cmd.Print("Local Changes:     ")
			if len(buildvars.GitPorcelain) == 0 {
				cmd.Print("false\n")
			} else {
				cmd.Printf("true\n%s\n", buildvars.GitPorcelain)
			}
		},
	}
}
