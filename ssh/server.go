package ssh

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
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
	usefullnessTimer := time.AfterFunc(5*time.Minute, func() {
		c.Close()
	})

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

		if channelRequest.ChannelType() != "direct-tcpip" {
			msg := fmt.Sprintf("no %s allowed, only direct-tcpip", channelRequest.ChannelType())
			channelRequest.Reject(ssh.Prohibited, msg)
			logger.Printf("%s: open illegal channel: %s", conn.RemoteAddr().String(), msg)
			continue
		}

		forwardInfo := directTCPIP{}
		err := ssh.Unmarshal(channelRequest.ExtraData(), &forwardInfo)
		if err != nil {
			logger.Printf("unable to unmarshal forward information: %s", err)
			channelRequest.Reject(ssh.UnknownChannelType, "failed to parse forward information")
			continue
		}

		// Accept channel from ssh client
		logger.Printf("accepting forward to %s:%d", forwardInfo.Addr, forwardInfo.Rport)
		_, requests, err := channelRequest.Accept()
		if err != nil {
			logger.Print("could not accept forward channel: ", err)
			continue
		}

		go ssh.DiscardRequests(requests)

	}

	logger.Print("client went away ", conn.RemoteAddr())
}
