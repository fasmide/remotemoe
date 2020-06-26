package http

import (
	"context"
	"net"
	"net/http"

	"github.com/fasmide/remotemoe/router"
	"golang.org/x/crypto/acme/autocert"
)

func NewServer(r *router.Router) *http.Server {
	m := &autocert.Manager{
		Cache:      autocert.DirCache("secret-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: r.Exists,
	}

	return &http.Server{
		ConnContext: withLocalAddr,
		TLSConfig:   m.TLSConfig(),
	}
}

type localAddr string

func withLocalAddr(ctx context.Context, c net.Conn) context.Context {
	ctx = context.WithValue(ctx, localAddr("localaddr"), c.LocalAddr().String())
	return ctx
}
