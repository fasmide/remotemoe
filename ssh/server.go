package ssh

import (
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
	session := Session{
		clearConn:       c,
		secureConn:      conn,
		channelRequests: chans,
		requests:        reqs,
	}

	session.Handle()

	logger.Print("client went away ", conn.RemoteAddr())
}
