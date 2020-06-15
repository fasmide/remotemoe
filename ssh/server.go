package ssh

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "[ssh] ", log.Flags())
}

// Server represents a listening ssh server
type Server struct {
	// Config is the ssh serverconfig
	Config *ssh.ServerConfig

	listener net.Listener
}

// Listen will listen and accept ssh connections
func (s *Server) Listen(l net.Listener) {
	s.listener = l
	for {
		nConn, err := s.listener.Accept()
		if err != nil {
			logger.Print("failed to accept incoming connection: ", err)
		}
		go s.accept(nConn)
	}
}

func (s *Server) accept(c net.Conn) {
	// auth timeout
	// only give people 10 seconds to ssh handshake and authenticate themselves
	authTimer := time.AfterFunc(10*time.Second, func() {
		c.Close()
	})

	// ssh handshake and auth
	conn, chans, reqs, err := ssh.NewServerConn(c, s.Config)
	if err != nil {
		logger.Print("failed to handshake: ", err)
		return
	}
	authTimer.Stop()

	logger.Printf("accepted session from %s", conn.RemoteAddr())

	// if a connection havnt done anything usefull within 5 minutes, throw them away
	// usefullnessTimer := time.AfterFunc(5*time.Minute, func() {
	// 	c.Close()
	// })

	// The incoming Request channel must be serviced.
	go func(reqs <-chan *ssh.Request) {
		for req := range reqs {
			if req.Type == "keepalive@openssh.com" {
				req.Reply(true, nil)
				continue
			}

			if req.Type == "tcpip-forward" {
				forwardInfo := tcpIPForward{}
				err := ssh.Unmarshal(req.Payload, &forwardInfo)

				if err != nil {
					logger.Printf("%s: unable to unmarshal forward information: %s", conn.RemoteAddr(), err)
					req.Reply(false, nil)
					continue
				}

				logger.Printf("%s: tcpip-forward: %+v", conn.RemoteAddr(), forwardInfo)
				req.Reply(true, nil)
				continue
			}

			logger.Printf("%s: unknown request-type: %s", conn.RemoteAddr(), req.Type)
			req.Reply(false, nil)

		}
	}(reqs)

	// Service the incoming Channel channel.
	for channelRequest := range chans {
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

	logger.Print("client went away ", conn.RemoteAddr())
}

// AcceptSession starts a new user terminal for the end user
func (s *Server) AcceptSession(session ssh.NewChannel) error {
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
	term := terminal.NewTerminal(channel, "> ")
	go func() {
		defer channel.Close()
		for {
			line, err := term.ReadLine()
			if err != nil {
				break
			}
			term.Write([]byte(fmt.Sprintf("What does %s mean?\r\n", line)))
		}
	}()

	return nil
}

// AcceptForwardRequest parses information about the request, checks to see if an endpoint
// matching is available in the router and then io.Copy'es everything back and forth
func (s *Server) AcceptForwardRequest(fr ssh.NewChannel) error {
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
