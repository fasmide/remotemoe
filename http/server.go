package http

import (
	"net"
	"net/http"
)

type Server struct {
	http.Server
}

func New() *Server {
	s := &Server{}
	return s
}

func (s *Server) Listen(l net.Listener) error {
	return s.Serve(l)
}
