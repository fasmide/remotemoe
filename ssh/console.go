package ssh

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
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

	// setup this sessions terminal
	term := term.NewTerminal(channel, "$ ")

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

			// the pty-req has information about the client terminal
			// we need the initial width and height of the terminal
			if req.Type == "pty-req" {
				ptyReq := ptyRequest{}
				err := ssh.Unmarshal(req.Payload, &ptyReq)
				if err != nil {
					log.Printf("unable to parse ssh pty-request: %s", err)
					req.Reply(false, nil)
					continue
				}

				term.SetSize(int(ptyReq.Width), int(ptyReq.Height))
				continue
			}

			// look for "window-change" requests - these should update the terminal size
			if req.Type == "window-change" {
				wcReq := windowChange{}
				err := ssh.Unmarshal(req.Payload, &wcReq)
				if err != nil {
					log.Printf("unable to parse ssh window-change request: %s", err)
					req.Reply(false, nil)
					continue
				}

				term.SetSize(int(wcReq.Width), int(wcReq.Height))
				continue
			}

			// reply false to everything other then shell
			req.Reply(req.Type == "shell", nil)
		}
	}(requests)

	// autocomplete and the actural command execution cannot access
	// the command at the same time
	var lock sync.Mutex
	main := DefaultCmd(c.session, c.session.router)
	main.SetOut(term)
	main.SetErr(term)

	term.AutoCompleteCallback = func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		lock.Lock()
		defer lock.Unlock()

		// If we don't receive TAB, simply return without doing anything.
		if key != '\t' {
			return line, pos, false
		}

		prefix := line[:pos]
		postfix := line[pos:]

		line = line[:pos] + string(key) + line[pos:]

		wordBeforeCursor := prefix

		spacePos := strings.LastIndex(prefix, " ")
		if spacePos >= 0 {
			wordBeforeCursor = prefix[spacePos+1:]
		}

		var suggestions []string

		command := main
		args := strings.Fields(line)

		if found, _, err := command.Find(args); err == nil {
			command = found
		}

		// If the cursor is placed at the end of a matched command, add a space. Like bash.
		if command.Name() == wordBeforeCursor {
			return prefix + " " + postfix, pos + 1, true
		}

		if command.HasAvailableSubCommands() {
			for _, c := range command.Commands() {
				name := c.Name()

				if !c.Hidden && strings.HasPrefix(name, wordBeforeCursor) {
					suggestions = append(suggestions, name)
				}
			}
		}

		// If we have exactly one match, simply use it.
		if len(suggestions) == 1 {
			// If the cursor is at the end, insert a space too like bash.
			if postfix == "" {
				postfix = " "
				pos++
			}

			return prefix + suggestions[0][len(wordBeforeCursor):] + postfix, pos + len(suggestions[0]) - len(wordBeforeCursor), true
		}

		if len(suggestions) > 1 {
			whiteSpace := ""
			for i := 0; i < len(prefix)+2; i++ {
				whiteSpace += " "
			}

			// Trick term into keeping the current line.
			fmt.Fprintf(term, "$ %s\033[33m...\033[0m%s\n\r", prefix, postfix)
			for _, s := range suggestions {
				fmt.Fprintf(term, "\033[33m%s%s\033[0m\n", whiteSpace, s)
			}
		}

		return prefix + postfix, pos, false
	}

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
					// the user should get a proper exit-code when closing the commandline manually
					channel.SendRequest("exit-status", false, ssh.Marshal(struct{ C uint32 }{0}))

					// break go routine
					return
				}

				if strings.TrimSpace(cmd) == "" {
					continue
				}

				lock.Lock()

				main.SetArgs(strings.Fields(cmd))
				_ = main.Execute()

				// the main commands needs its flags reverted to their default value
				// for every invocation
				CommandReset(main)

				lock.Unlock()
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
