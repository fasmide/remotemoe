package ssh

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/fasmide/remotemoe/router"
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

	// Router is used to expose routes
	Router *router.Router

	listener net.Listener
}

// Serve will accept ssh connections
func (s *Server) Serve(l net.Listener) error {
	s.listener = l
	for {
		nConn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept incoming connection: %w", err)
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
		router:          s.Router,
	}

	session.Handle()

	logger.Print("client went away ", conn.RemoteAddr())
}
