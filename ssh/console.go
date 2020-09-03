package ssh

import (
	"fmt"
	"io"
	"log"

	"github.com/cosiner/argv"
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// Console represents a terminal and handles commands
type Console struct {
	session *Session
}

// Accept accepts a session NewChannel and sets up the terminal and commands
func (c *Console) Accept(channelRequest ssh.NewChannel) error {
	channel, requests, err := channelRequest.Accept()
	if err != nil {
		return fmt.Errorf("unable to accept channel: %w", err)
	}

	// commands comes from "exec" requests or when a user enters them into the shell
	commands := make(chan string)

	// reply "success" to shell and pty-req's
	go func(in <-chan *ssh.Request) {
		for req := range in {
			if req.Type == "exec" {
				// parse exec request
				exec := execCommand{}
				err := ssh.Unmarshal(req.Payload, &exec)
				if err != nil {
					log.Printf("unable to parse exec payload: %s", err)
					req.Reply(false, nil)
					continue
				}

				// queue command which will be executed later
				// when the client opens a shell
				commands <- exec.Command
				continue
			}

			// reply false to everything other then shell and pty-req
			req.Reply(req.Type == "shell" || req.Type == "pty-req", nil)
		}
	}(requests)

	// setup this sessions terminal
	term := terminal.NewTerminal(channel, "$ ")

	// read commands off the terminal and put them into commands channel
	go func() {
		for {
			line, err := term.ReadLine()
			if err != nil {
				close(commands)
				break
			}
			commands <- line
		}
	}()

	fmt.Fprintf(term, "New to remotemoe? - try 'firsttime' or 'help' and start exploring!\r\n\r\n")

	go func() {
		defer channel.Close()
		for {
			select {
			case cmd, ok := <-commands:
				if !ok {
					return
				}

				// args will seperate command arvg1 | anothercommand argv1 and put the result into
				// [][]string
				args, err := argv.Argv(cmd, func(backquoted string) (string, error) {
					return backquoted, nil
				}, nil)

				if err != nil {
					fmt.Fprintf(term, "failed to parse: %s\r\n", err)
					continue
				}

				for _, cmd := range args {
					c.handleCommand(cmd, term)
				}

			case msg, ok := <-c.session.msgs:
				if !ok {
					return
				}
				fmt.Fprintf(term, "%s\r\n", msg)
			}

			c.session.PokeTimeout()
		}
	}()

	return nil
}

func (c *Console) handleCommand(argv []string, output io.Writer) {
	bold := color.New(color.Bold)

	// forces colors on
	bold.EnableColor()

	if len(argv) == 0 {
		return
	}

	switch argv[0] {
	case "exit":
		fmt.Fprint(output, "bye!\r\n")
		c.session.Close()
	case "quit":
		fmt.Fprint(output, "bye!\r\n")
		c.session.Close()
	case "coffie":
		fmt.Fprint(output, "Sure! - have some coffie\r\n")
	case "help":
		fmt.Fprint(output, bold.Sprint("Commands:"))
		fmt.Fprint(output, "\r\n\r\n")

		fmt.Fprint(output, "  services    info about services currently active\r\n")
		fmt.Fprint(output, "  add         add hostname(s) to your services\r\n")
		fmt.Fprint(output, "  remove      remove hostname(s) from your services\r\n")
		fmt.Fprint(output, "  remove all  removes all hostnames from your services\r\n")

		fmt.Fprintf(output, "\r\n%s\r\n\r\n", bold.Sprint("Ways of keeping an ssh connection open:"))
		fmt.Fprint(output, "  autossh     using autossh\r\n")
		fmt.Fprint(output, "  unitfile    using a systemd unit\r\n")
		fmt.Fprint(output, "  bashloop    using a simple bash loop\r\n")

		fmt.Fprintf(output, "\r\n%s\r\n\r\n", bold.Sprint("Other topics:"))
		fmt.Fprint(output, "  firsttime   first time users of remotemoe and ssh tunneling\r\n")
		fmt.Fprint(output, "  forwards    intro to ssh forward ports with `-R`\r\n")

		fmt.Fprint(output, "\r\n")
	case "remove":
		if len(argv) == 1 {
			fmt.Fprintf(output,
				"%s:\r\n\r\n  remove some-hostname.%s [more-hostnames.%s] [my.very.own.domain.com] ... \r\n\r\n",
				bold.Sprintf("remove usage"),
				services.Hostname,
				services.Hostname,
			)

			fmt.Fprint(output, "Remove hostname(s) previously added.\r\n")
			fmt.Fprintf(output, "Tip: %s will remove every named route owned by you.\r\n", bold.Sprintf("remove all"))
			return
		}

		if argv[1] == "all" {
			removed, err := router.RemoveAll(c.session)
			if err != nil {
				fmt.Fprintf(output, "could not remove all names: %s\r\n", err)
				fmt.Fprint(output, "some may have been removed.\r\n")
				return
			}

			for _, nr := range removed {
				fmt.Fprintf(output, "%s removed.\r\n", bold.Sprint(nr.FQDN()))
			}
			return
		}

		// remove all provided names
		for _, name := range argv[1:] {
			err := router.RemoveName(name, c.session)
			if err != nil {
				fmt.Fprintf(output, "could not remove %s: %s\r\n", bold.Sprintf(name), err)
				continue
			}

			fmt.Fprintf(output, "%s removed.\r\n", bold.Sprintf(name))
		}

	case "add":
		if len(argv) == 1 {
			fmt.Fprintf(output,
				"%s:\r\n\r\n  add some-wanted-hostname.%s [more-hostnames.%s] [my.very.own.domain.com] ... \r\n\r\n",
				bold.Sprintf("add usage"),
				services.Hostname,
				services.Hostname,
			)
			fmt.Fprint(output, "Add as many hostnames as needed. You can bring your own domains by setting up DNS records appropriately.\r\n\r\n")
			fmt.Fprintf(output, "Check out %s command for all active hostnames\r\n", bold.Sprint("services"))
			return
		}

		for _, n := range argv[1:] {
			namedRoute := router.NewName(n, c.session)
			err := router.Add(namedRoute)
			if err != nil {
				fmt.Fprintf(output, "%s could not be added: %s\r\n", bold.Sprint(n), err)
				continue
			}
			fmt.Fprintf(output, "%s is active.\r\n", bold.Sprint(namedRoute.FQDN()))
		}

	case "autossh":
		fmt.Fprintf(output,
			"# autossh template based on ports %s\r\n",
			bold.Sprint(joinDigits(c.session.serviceKeys())),
		)
		fmt.Fprint(output, "autossh -M 0 -f \\\r\n")
		fmt.Fprint(output, "  -o \"ExitOnForwardFailure yes\" \\\r\n")
		fmt.Fprint(output, "  -o \"ServerAliveInterval 30\" \\\r\n")
		fmt.Fprint(output, "  -o \"ServerAliveCountMax 3\" \\\r\n")

		for p := range c.session.services {
			fmt.Fprintf(output, "  -R %d:localhost:%d \\\r\n", p, p)
		}

		fmt.Fprintf(output, "  %s -N\r\n", services.Hostname)
		fmt.Fprint(output, "\r\n")
		fmt.Fprint(output, "# for this to work, autossh needs access to the same keys and known_hosts as you had.\r\n")
		fmt.Fprint(output, "# if debugging is needed, remove the `-f` parameter which will keep autossh in the foreground.\r\n")
		fmt.Fprint(output, "\r\n")
	case "unitfile":
		fmt.Fprint(output, bold.Sprintf("systemd user service unit"), "\r\n")
		fmt.Fprint(output, "Ensure you have lingering turned on, and the directories setup correctly:\r\n")
		fmt.Fprint(output, "  $ mkdir -p ~/.config/systemd/user/\r\n")
		fmt.Fprint(output, "  $ sudo loginctl enable-linger $USER\r\n")
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "Put this file, into %s\r\n", bold.Sprintf("~/.config/systemd/user/remotemoe.service"))
		fmt.Fprint(output, "[Unit]\r\nDescription=remotemoe tunnel\r\nStartLimitIntervalSec=0\r\nAfter=network.target\r\n\r\n[Service]\r\nRestart=always\r\nRestartSec=1m\r\n")
		fmt.Fprint(output, "ExecStart=ssh \\\r\n")
		fmt.Fprint(output, "  -o \"ExitOnForwardFailure yes\" \\\r\n")
		fmt.Fprint(output, "  -o \"ServerAliveInterval 30\"  \\\r\n")
		fmt.Fprint(output, "  -o \"ServerAliveCountMax 3\" \\\r\n")

		for p := range c.session.services {
			fmt.Fprintf(output, "  -R %d:localhost:%d \\\r\n", p, p)
		}

		fmt.Fprintf(output, "  %s -N\r\n", services.Hostname)
		fmt.Fprint(output, "\r\n")
		fmt.Fprint(output, "[Install]\r\nWantedBy=default.target\r\n")
		fmt.Fprint(output, "\r\n")
		fmt.Fprint(output, "You should now be able to start the service:\r\n")
		fmt.Fprint(output, " $ systemctl --user start remotemoe.service\r\n")
		fmt.Fprint(output, "\r\n")
		fmt.Fprint(output, "You can also enable the service at boot time:\r\n")
		fmt.Fprint(output, " $ systemctl --user enable remotemoe.service\r\n")
		fmt.Fprint(output, "\r\n")

	case "bashloop":
		fmt.Fprint(output, "FIXME: Here be bash loop\r\n")
	case "firsttime":
		fmt.Fprintf(output, "%s\r\n", bold.Sprintf("remotemoe"))
		fmt.Fprint(output, "remotemoe allows users to access services that are otherwise inaccessible from the internet.\r\n")
		fmt.Fprint(output, "Just like ngrok or argo tunnels, a device or service connects to remotemoe which in turn muxes ")
		fmt.Fprint(output, "requests back from the internet. \r\n\r\n")

		fmt.Fprintf(output, "%s\r\n", bold.Sprintf("Basic example:"))
		fmt.Fprint(output, "Access the command line and a webservice of a remotely deployed Raspberry Pi:\r\n\r")

		fmt.Fprint(output, firstTimeDiagram)

		fmt.Fprint(output, "\r\n\r\n")
		fmt.Fprint(output, "From the Raspberry pi, connect using `-R` parameters which tells ssh to forward ports.")
		fmt.Fprint(output, "\r\n\r\n")
		fmt.Fprintf(output, "  ssh -R80:localhost:80 -R22:localhost:22 %s\r\n\r\n", services.Hostname)
		fmt.Fprint(output, "That's it, the Raspberry Pi's webservice and ssh daemon are now accessible from the internet\r\n")
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "For information on how to access the services, have a look at the %s command\r\n", bold.Sprintf("services"))
	case "forwards":
		fmt.Fprint(output, "First off, take a look in the ssh(1) manual and look for the `-R` parameter.\r\n\r\n")
		fmt.Fprint(output, "remotemoe uses the ports and hostnames like this:\r\n")
		fmt.Fprint(output, forwardDiagram)
		fmt.Fprint(output, "\r\n")
		fmt.Fprint(output, "Incoming port forwards are mapped directly to service mux'es, with the following rules:\r\n")
		fmt.Fprintf(output, "Ports %s will be accessible with %s\r\n", bold.Sprint(joinDigits(services.Services["http"])), bold.Sprint("HTTP"))
		fmt.Fprintf(output, "Ports %s will be accessible with %s\r\n", bold.Sprint(joinDigits(services.Services["https"])), bold.Sprint("HTTPs"))
		fmt.Fprintf(output, "Ports %s will be accessible with %s\r\n", bold.Sprint(joinDigits(services.Services["ssh"])), bold.Sprint("ssh"))
	case "services":
		namedRoutes, err := router.Names(c.session)
		if err != nil {
			fmt.Fprint(output, "unable to lookup your custom names, try again later...\r\n")
			// we should let the command continue but with an empty slice
			namedRoutes = make([]router.NamedRoute, 0, 0)
		}

		// Write a few sentences about currently forwarded ports...
		if len(c.session.services) == 0 {
			fmt.Fprintf(output, "You have %s forwarded ports, have a look in the ssh manual: %s.\r\n", bold.Sprint("zero"), bold.Sprint("man ssh"))
			fmt.Fprintf(output, "You will be looking for the %s parameter.\r\n", bold.Sprint("-R"))
		} else {
			fmt.Fprintf(output,
				"Based on currently forwarded ports %s, your services will be available at:\r\n",
				bold.Sprint(joinDigits(c.session.serviceKeys())),
			)
		}

		// HTTP services
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "%s (%s)", bold.Sprint("HTTP"), joinDigits(services.Services["http"]))
		fmt.Fprint(output, "\r\n")

		help := true
		for _, p := range services.Services["http"] {
			if _, exists := c.session.services[uint32(p)]; !exists {
				continue
			}

			// do not display further help about http ports
			help = false

			// port 80 being the default http port - omit the :port format
			if p == 80 {
				fmt.Fprintf(output, "http://%s/\r\n", c.session.FQDN())
				for _, nr := range namedRoutes {
					fmt.Fprintf(output, "http://%s/\r\n", nr.FQDN())
				}
				continue
			}

			fmt.Fprintf(output, "http://%s:%d/\r\n", c.session.FQDN(), p)
			for _, nr := range namedRoutes {
				fmt.Fprintf(output, "http://%s:%d/\r\n", nr.FQDN(), p)
			}
		}

		if help {
			fmt.Fprintf(output, "No HTTP services found, add some by appending `-R80:localhost:80` when connecting.\r\n")
		}

		// HTTPS services
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "%s (%s)", bold.Sprint("HTTPS"), joinDigits(services.Services["https"]))
		fmt.Fprint(output, "\r\n")

		help = true
		for _, p := range services.Services["https"] {
			if _, exists := c.session.services[uint32(p)]; !exists {
				continue
			}

			// do not display further help about https ports
			help = false

			// port 443 being the default http port - omit the :port format
			if p == 443 {
				fmt.Fprintf(output, "https://%s/\r\n", c.session.FQDN())
				for _, nr := range namedRoutes {
					fmt.Fprintf(output, "https://%s/\r\n", nr.FQDN())
				}
				continue
			}

			fmt.Fprintf(output, "https://%s:%d/\r\n", c.session.FQDN(), p)
			for _, nr := range namedRoutes {
				fmt.Fprintf(output, "https://%s:%d/\r\n", nr.FQDN(), p)
			}
		}

		if help {
			fmt.Fprintf(output, "No HTTPS services found, add some by appending `-R443:localhost:443` when connecting.\r\n")
		}

		// SSH services
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "%s (%s)", bold.Sprint("SSH"), joinDigits(services.Services["ssh"]))
		fmt.Fprint(output, "\r\n")

		help = true
		for _, p := range services.Services["ssh"] {
			if _, exists := c.session.services[uint32(p)]; !exists {
				continue
			}

			// do not display further help about ssh ports
			help = false

			// port 22 being the default ssh port - omit the -p<port> format
			if p == 22 {
				fmt.Fprintf(output, "ssh -J %s %s\r\n", services.Hostname, c.session.FQDN())
				for _, nr := range namedRoutes {
					fmt.Fprintf(output, "ssh -J %s %s\r\n", services.Hostname, nr.FQDN())
				}
				continue
			}

			fmt.Fprintf(output, "ssh -p%d -J %s:%d %s\r\n", p, services.Hostname, p, c.session.FQDN())
			for _, nr := range namedRoutes {
				fmt.Fprintf(output, "ssh -p%d -J %s:%d %s\r\n", p, services.Hostname, p, nr.FQDN())

			}
		}

		if help {
			fmt.Fprintf(output, "No SSH services found, add some by appending `-R22:localhost:22` when connecting.\r\n")
		}

	default:
		fmt.Fprintf(output, "%s: command not found\r\n", argv[0])
	}
}
