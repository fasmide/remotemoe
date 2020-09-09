package command

import (
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

// ForwardingRoutable is an interface for things that are routable and provides forwards
type ForwardingRoutable interface {
	router.Routable
	Forwarding
}

// Access returns a *cobra.Command which tells the user how to access services
func Access(fr ForwardingRoutable) *cobra.Command {
	return &cobra.Command{
		Use:   "access",
		Short: "How to access forwarded services",
		Run: func(cmd *cobra.Command, _ []string) {
			namedRoutes, err := router.Names(fr)
			if err != nil {
				cmd.Printf("unable to lookup your custom names, try again later...\n")
				// we should let the command continue but with an empty slice
				namedRoutes = make([]router.NamedRoute, 0, 0)
			}

			forwards := fr.Forwards()

			// Write a few sentences about currently forwarded ports...
			if len(forwards) == 0 {
				cmd.Printf("You have %s forwarded ports, have a look in the ssh manual: %s.\n", "zero", "man ssh")
				cmd.Printf("You will be looking for the %s parameter.\n", "-R")
			} else {
				cmd.Printf(
					"Based on currently forwarded ports %s, your services will be available at:\n",
					joinDigits(serviceKeys(forwards)),
				)
			}

			// HTTP services
			cmd.Print("\n")
			cmd.Printf("%s (%s)", "HTTP", joinDigits(services.Services["http"]))
			cmd.Print("\n")

			help := true

			for _, p := range services.Services["http"] {
				if _, exists := forwards[uint32(p)]; !exists {
					continue
				}

				// do not display further help about http ports
				help = false

				// port 80 being the default http port - omit the :port format
				if p == 80 {
					cmd.Printf("http://%s/\n", fr.FQDN())
					for _, nr := range namedRoutes {
						cmd.Printf("http://%s/\n", nr.FQDN())
					}
					continue
				}

				cmd.Printf("http://%s:%d/\n", fr.FQDN(), p)
				for _, nr := range namedRoutes {
					cmd.Printf("http://%s:%d/\n", nr.FQDN(), p)
				}
			}

			if help {
				cmd.Printf("No HTTP services found, add some by appending `-R80:localhost:80` when connecting.\n")
			}

			// HTTPS services
			cmd.Print("\n")
			cmd.Printf("%s (%s)", "HTTPS", joinDigits(services.Services["https"]))
			cmd.Print("\n")

			help = true

			for _, p := range services.Services["https"] {
				if _, exists := forwards[uint32(p)]; !exists {
					continue
				}

				// do not display further help about https ports
				help = false

				// port 443 being the default http port - omit the :port format
				if p == 443 {
					cmd.Printf("https://%s/\n", fr.FQDN())
					for _, nr := range namedRoutes {
						cmd.Printf("https://%s/\n", nr.FQDN())
					}
					continue
				}

				cmd.Printf("https://%s:%d/\n", fr.FQDN(), p)
				for _, nr := range namedRoutes {
					cmd.Printf("https://%s:%d/\n", nr.FQDN(), p)
				}
			}

			if help {
				cmd.Printf("No HTTPS services found, add some by appending `-R443:localhost:443` when connecting.\n")
			}

			// SSH services
			cmd.Print("\n")
			cmd.Printf("%s (%s)", "SSH", joinDigits(services.Services["ssh"]))
			cmd.Print("\n")

			help = true

			for _, p := range services.Services["ssh"] {
				if _, exists := forwards[uint32(p)]; !exists {
					continue
				}

				// do not display further help about ssh ports
				help = false

				// port 22 being the default ssh port - omit the -p<port> format
				if p == 22 {
					cmd.Printf("ssh -J %s %s\n", services.Hostname, fr.FQDN())
					for _, nr := range namedRoutes {
						cmd.Printf("ssh -J %s %s\n", services.Hostname, nr.FQDN())
					}
					continue
				}

				cmd.Printf("ssh -p%d -J %s:%d %s\n", p, services.Hostname, p, fr.FQDN())
				for _, nr := range namedRoutes {
					cmd.Printf("ssh -p%d -J %s:%d %s\n", p, services.Hostname, p, nr.FQDN())

				}
			}

			if help {
				cmd.Printf("No SSH services found, add some by appending `-R22:localhost:22` when connecting.\n")
			}

		},
	}
}
