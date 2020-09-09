package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

// Autossh returns a *cobra.Command which generates an autossh snippit
func Autossh(d Forwarding) *cobra.Command {
	return &cobra.Command{
		Use:   "autossh",
		Short: "Generates snippet for use with autossh",
		Run: func(cmd *cobra.Command, args []string) {
			ports := d.Forwards()
			cmd.Printf(
				"# autossh template based on ports %s\r\n",
				joinDigits(serviceKeys(ports)),
			)

			cmd.Print("autossh -M 0 -f \\\n")
			cmd.Print("  -o \"ExitOnForwardFailure yes\" \\\n")
			cmd.Print("  -o \"ServerAliveInterval 30\" \\\n")
			cmd.Print("  -o \"ServerAliveCountMax 3\" \\\n")

			for p := range ports {
				cmd.Printf("  -R %d:localhost:%d \\\n", p, p)
			}

			cmd.Printf("  %s -N\r\n", services.Hostname)
			cmd.Print("\n")
			cmd.Print("# for this to work, autossh needs access to the same keys and known_hosts as you had.\n")
			cmd.Print("# if debugging is needed, remove the `-f` parameter which will keep autossh in the foreground.\n")
			cmd.Print("\n")
		},
	}
}

func serviceKeys(s map[uint32]struct{}) []int {
	keys := make([]int, 0, len(s))
	for v := range s {
		keys = append(keys, int(v))
	}
	sort.Sort(sort.IntSlice(keys))
	return keys
}

func joinDigits(ds []int) string {
	b := &strings.Builder{}
	for i, v := range ds {
		if i == 0 {
			fmt.Fprintf(b, "%d", v)
			continue
		}
		if i == len(ds)-1 {
			fmt.Fprintf(b, " and %d", v)
			continue
		}
		fmt.Fprintf(b, ", %d", v)
	}
	return b.String()
}
