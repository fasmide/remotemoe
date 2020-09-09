package command

import (
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

// Unitfile returns a cobra.Command that generates a systemd unit file based on the current session
func Unitfile(d Forwarding) *cobra.Command {
	return &cobra.Command{
		Use:   "unitfile",
		Short: "Generates an unitfile for use with systemd",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print("systemd user service unit\n")
			cmd.Print("Ensure you have lingering turned on, and the directories setup correctly:\n")
			cmd.Print("  $ mkdir -p ~/.config/systemd/user/\n")
			cmd.Print("  $ sudo loginctl enable-linger $USER\n")
			cmd.Print("\n")
			cmd.Printf("Put this file, into ~/.config/systemd/user/remotemoe.service\n")
			cmd.Print("[Unit]\nDescription=remotemoe tunnel\nStartLimitIntervalSec=0\nAfter=network.target\n\n[Service]\nRestart=always\nRestartSec=1m\n")
			cmd.Print("ExecStart=ssh \\\n")
			cmd.Print("  -o \"ExitOnForwardFailure yes\" \\\n")
			cmd.Print("  -o \"ServerAliveInterval 30\"  \\\n")
			cmd.Print("  -o \"ServerAliveCountMax 3\" \\\n")

			for p := range d.Forwards() {
				cmd.Printf("  -R %d:localhost:%d \\\n", p, p)
			}

			cmd.Printf("  %s -N\n", services.Hostname)
			cmd.Print("\n")
			cmd.Print("[Install]\nWantedBy=default.target\n")
			cmd.Print("\n")
			cmd.Print("You should now be able to start the service:\n")
			cmd.Print(" $ systemctl --user start remotemoe.service\n")
			cmd.Print("\n")
			cmd.Print("You can also enable the service at boot time:\n")
			cmd.Print(" $ systemctl --user enable remotemoe.service\n")
			cmd.Print("\n")
		},
	}
}
