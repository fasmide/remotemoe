package http

import (
	"context"
	"net"
	"net/http"
)

func NewServer() *http.Server {
	return &http.Server{
		ConnContext: withLocalAddr,
	}
}

type localAddr string

func withLocalAddr(ctx context.Context, c net.Conn) context.Context {
	ctx = context.WithValue(ctx, localAddr("localaddr"), c.LocalAddr().String())
	return ctx
}
