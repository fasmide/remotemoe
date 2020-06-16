package ssh

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// IdleTimeout sets how long a session can be idle before getting disconnected
const IdleTimeout = time.Minute

// Session represents a ongoing SSH connection
type Session struct {
	// Raw socket
	clearConn net.Conn
	// SSH Socket
	secureConn *ssh.ServerConn
	// Incoming channel requests
	channelRequests <-chan ssh.NewChannel
	// Global requests
	requests <-chan *ssh.Request

	idleLock     sync.Mutex
	idleTimeout  *time.Timer
	idleDisabled bool

	// messages to the terminal (i.e. the user)
	msgs chan string
}

func (s *Session) Handle() {

	// initialize msgs channel
	// we are doing a buffered channel, as a slutty way of not blocking `-N` connections
	// as no terminal is available, we will just buffer them and
	// get on with our lives ... time will tell if this is a good idea :)
	s.msgs = make(chan string, 15)

	// if a connection havnt done anything usefull within a minute, throw them away
	s.idleTimeout = time.AfterFunc(IdleTimeout, s.Timeout)

	// The incoming Request channel must be serviced.
	go s.HandleRequests()

	// block here until the end of time
	s.HandleChannels()

	// No reason to keep the timer active
	s.DisableTimeout()
}

// Timeout fires when the session has done too much idling
func (s *Session) Timeout() {
	logger.Printf("%s idle for more then %s:, closing", s.clearConn.RemoteAddr(), IdleTimeout)

	err := s.secureConn.Close()
	if err != nil {
		logger.Printf("could not close secureConnection: %s", err)
	}

}

// PokeTimeout adds to its duration - unless disabled
func (s *Session) PokeTimeout() {
	s.idleLock.Lock()
	defer s.idleLock.Unlock()

	if s.idleDisabled {
		return
	}

	if !s.idleTimeout.Stop() {
		// too late - it already fired
		return
	}

	s.idleTimeout.Reset(IdleTimeout)
}

// DisableTimeout disables the idle timeout, used when a connection provides some endpoints
// i.e. requests ports to be forwarded...
func (s *Session) DisableTimeout() {
	s.idleLock.Lock()
	defer s.idleLock.Unlock()

	s.idleTimeout.Stop()
	s.idleDisabled = true
}

func (s *Session) HandleChannels() {
	for channelRequest := range s.channelRequests {
		// direct-tcpip forward requests
		if channelRequest.ChannelType() == "direct-tcpip" {
			err := s.AcceptForwardRequest(channelRequest)
			if err != nil {
				logger.Printf("unable to accept channel: %s", err)
			}
			continue
		}

		if channelRequest.ChannelType() == "session" {
			err := s.AcceptSession(channelRequest)
			if err != nil {
				logger.Printf("unable to accept session: %s", err)
			}
			continue
		}

	}
}

func (s *Session) HandleRequests() {

	for req := range s.requests {
		if req.Type == "keepalive@openssh.com" {
			req.Reply(true, nil)
			continue
		}

		if req.Type == "tcpip-forward" {
			forwardInfo := tcpIPForward{}
			err := ssh.Unmarshal(req.Payload, &forwardInfo)

			if err != nil {
				logger.Printf("%s: unable to unmarshal forward information: %s", s.clearConn.RemoteAddr(), err)
				req.Reply(false, nil)
				continue
			}

			logger.Printf("%s: tcpip-forward: %+v", s.clearConn.RemoteAddr(), forwardInfo)
			req.Reply(true, nil)

			s.msgs <- fmt.Sprintf("I have accepted your port %d, it will be available on %s", forwardInfo.Rport, "lkasjdlkajsd.eu.remote.moe")
			s.DisableTimeout()
			continue
		}

		logger.Printf("%s: unknown request-type: %s", s.clearConn.RemoteAddr(), req.Type)
		req.Reply(false, nil)

	}

}

// AcceptSession starts a new user terminal for the end user
func (s *Session) AcceptSession(session ssh.NewChannel) error {
	channel, requests, err := session.Accept()
	if err != nil {
		return fmt.Errorf("unable to accept channel: %w", err)
	}

	// reply "success" to shell and pty-req's
	go func(in <-chan *ssh.Request) {
		for req := range in {
			req.Reply(req.Type == "shell" || req.Type == "pty-req", nil)
		}
	}(requests)

	// setup this sessions terminal
	fmt.Fprintf(channel, "Hello\r\nThis is remotemoe - take a look around...\r\n")
	term := terminal.NewTerminal(channel, "$ ")

	// read commands off the terminal and put them into commands channel
	commands := make(chan string)
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
				s.HandleCommand(cmd, term)
			case msg, ok := <-s.msgs:
				if !ok {
					return
				}
				fmt.Fprintf(term, "%s\r\n", msg)
			}
			s.PokeTimeout()
		}
	}()

	return nil
}

func (s *Session) HandleCommand(c string, output io.Writer) {
	switch c {
	case "":
		// nothing
	case "coffie":
		fmt.Fprint(output, "Sure! - have some coffie\r\n")
	default:
		fmt.Fprintf(output, "%s: command not found\r\n", c)
	}
}

// AcceptForwardRequest parses information about the request, checks to see if an endpoint
// matching is available in the router and then io.Copy'es everything back and forth
func (s *Session) AcceptForwardRequest(fr ssh.NewChannel) error {
	forwardInfo := directTCPIP{}
	err := ssh.Unmarshal(fr.ExtraData(), &forwardInfo)
	if err != nil {
		fr.Reject(ssh.UnknownChannelType, "failed to parse forward information")
		return fmt.Errorf("unable to unmarshal forward information: %w", err)
	}

	// Accept channel from ssh client
	logger.Printf("accepting forward to %s:%d", forwardInfo.Addr, forwardInfo.Rport)
	_, requests, err := fr.Accept()
	if err != nil {
		return fmt.Errorf("could not accept forward channel: %w", err)
	}

	// lookup "hostname" in the router, fetch remote and pass data

	go ssh.DiscardRequests(requests)

	return nil
}
