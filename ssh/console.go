package ssh

import (
	"fmt"
	"log"
	"strings"

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

	// handle "shell", "pty-req" and "exec" requests
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

	// DefaultCmd is our top level command, which embedds all others
	main := DefaultCmd(c.session)
	main.SetOut(term)

	go func() {
		defer channel.Close()
		for {
			select {
			case cmd, ok := <-commands:
				if !ok {
					return
				}

				if strings.TrimSpace(cmd) != "" {
					main.SetArgs(strings.Fields(cmd))
					_ = main.Execute()
				}

			case msg, ok := <-c.session.msgs:
				if !ok {
					return
				}
				fmt.Fprintf(term, "%s\n", msg)
			}

			c.session.PokeTimeout()
		}
	}()

	return nil
}

// func (c *Console) handleCommand(argv []string, output io.Writer) {
// 	bold := color.New(color.Bold)

// 	// forces colors on
// 	bold.EnableColor()

// 	if len(argv) == 0 {
// 		return
// 	}

// 	switch argv[0] {
// 	case "services":
// 		namedRoutes, err := router.Names(c.session)
// 		if err != nil {
// 			fmt.Fprint(output, "unable to lookup your custom names, try again later...\r\n")
// 			// we should let the command continue but with an empty slice
// 			namedRoutes = make([]router.NamedRoute, 0, 0)
// 		}

// 		// Write a few sentences about currently forwarded ports...
// 		if len(c.session.services) == 0 {
// 			fmt.Fprintf(output, "You have %s forwarded ports, have a look in the ssh manual: %s.\r\n", bold.Sprint("zero"), bold.Sprint("man ssh"))
// 			fmt.Fprintf(output, "You will be looking for the %s parameter.\r\n", bold.Sprint("-R"))
// 		} else {
// 			fmt.Fprintf(output,
// 				"Based on currently forwarded ports %s, your services will be available at:\r\n",
// 				bold.Sprint(joinDigits(c.session.serviceKeys())),
// 			)
// 		}

// 		// HTTP services
// 		fmt.Fprint(output, "\r\n")
// 		fmt.Fprintf(output, "%s (%s)", bold.Sprint("HTTP"), joinDigits(services.Services["http"]))
// 		fmt.Fprint(output, "\r\n")

// 		help := true
// 		for _, p := range services.Services["http"] {
// 			if _, exists := c.session.services[uint32(p)]; !exists {
// 				continue
// 			}

// 			// do not display further help about http ports
// 			help = false

// 			// port 80 being the default http port - omit the :port format
// 			if p == 80 {
// 				fmt.Fprintf(output, "http://%s/\r\n", c.session.FQDN())
// 				for _, nr := range namedRoutes {
// 					fmt.Fprintf(output, "http://%s/\r\n", nr.FQDN())
// 				}
// 				continue
// 			}

// 			fmt.Fprintf(output, "http://%s:%d/\r\n", c.session.FQDN(), p)
// 			for _, nr := range namedRoutes {
// 				fmt.Fprintf(output, "http://%s:%d/\r\n", nr.FQDN(), p)
// 			}
// 		}

// 		if help {
// 			fmt.Fprintf(output, "No HTTP services found, add some by appending `-R80:localhost:80` when connecting.\r\n")
// 		}

// 		// HTTPS services
// 		fmt.Fprint(output, "\r\n")
// 		fmt.Fprintf(output, "%s (%s)", bold.Sprint("HTTPS"), joinDigits(services.Services["https"]))
// 		fmt.Fprint(output, "\r\n")

// 		help = true
// 		for _, p := range services.Services["https"] {
// 			if _, exists := c.session.services[uint32(p)]; !exists {
// 				continue
// 			}

// 			// do not display further help about https ports
// 			help = false

// 			// port 443 being the default http port - omit the :port format
// 			if p == 443 {
// 				fmt.Fprintf(output, "https://%s/\r\n", c.session.FQDN())
// 				for _, nr := range namedRoutes {
// 					fmt.Fprintf(output, "https://%s/\r\n", nr.FQDN())
// 				}
// 				continue
// 			}

// 			fmt.Fprintf(output, "https://%s:%d/\r\n", c.session.FQDN(), p)
// 			for _, nr := range namedRoutes {
// 				fmt.Fprintf(output, "https://%s:%d/\r\n", nr.FQDN(), p)
// 			}
// 		}

// 		if help {
// 			fmt.Fprintf(output, "No HTTPS services found, add some by appending `-R443:localhost:443` when connecting.\r\n")
// 		}

// 		// SSH services
// 		fmt.Fprint(output, "\r\n")
// 		fmt.Fprintf(output, "%s (%s)", bold.Sprint("SSH"), joinDigits(services.Services["ssh"]))
// 		fmt.Fprint(output, "\r\n")

// 		help = true
// 		for _, p := range services.Services["ssh"] {
// 			if _, exists := c.session.services[uint32(p)]; !exists {
// 				continue
// 			}

// 			// do not display further help about ssh ports
// 			help = false

// 			// port 22 being the default ssh port - omit the -p<port> format
// 			if p == 22 {
// 				fmt.Fprintf(output, "ssh -J %s %s\r\n", services.Hostname, c.session.FQDN())
// 				for _, nr := range namedRoutes {
// 					fmt.Fprintf(output, "ssh -J %s %s\r\n", services.Hostname, nr.FQDN())
// 				}
// 				continue
// 			}

// 			fmt.Fprintf(output, "ssh -p%d -J %s:%d %s\r\n", p, services.Hostname, p, c.session.FQDN())
// 			for _, nr := range namedRoutes {
// 				fmt.Fprintf(output, "ssh -p%d -J %s:%d %s\r\n", p, services.Hostname, p, nr.FQDN())

// 			}
// 		}

// 		if help {
// 			fmt.Fprintf(output, "No SSH services found, add some by appending `-R22:localhost:22` when connecting.\r\n")
// 		}

// 	default:
// 		fmt.Fprintf(output, "%s: command not found\r\n", argv[0])
// 	}
// }
