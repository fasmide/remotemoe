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

	go func() {
		defer channel.Close()
		for {
			select {
			case cmd, ok := <-commands:
				if !ok {
					return
				}

				if strings.TrimSpace(cmd) != "" {
					main := DefaultCmd(c.session)
					main.SetOut(term)

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
