package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/fasmide/remotemoe/router"
	"golang.org/x/crypto/acme/autocert"
)

// NewServer returns a HTTP(S) capable server
func NewServer() (*http.Server, error) {
	cache, err := acmeCache()
	if err != nil {
		return nil, fmt.Errorf("unable to get acme cache: %w", err)
	}

	m := &autocert.Manager{
		Cache:      cache,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: router.Exists,
	}

	return &http.Server{
		ConnContext: withLocalAddr,
		TLSConfig:   m.TLSConfig(),
	}, nil
}

type localAddr string

func withLocalAddr(ctx context.Context, c net.Conn) context.Context {
	ctx = context.WithValue(ctx, localAddr("localaddr"), c.LocalAddr().String())
	return ctx
}

// acmeCache tries to find a systemd created state directory
// and oterwise defaults to $(pwd)/acme-secrets
func acmeCache() (autocert.Cache, error) {
	dir := "acme-secrets"

	if os.Getenv("STATE_DIRECTORY") != "" {
		dir = path.Join(os.Getenv("STATE_DIRECTORY"), "acme-secrets")
	}

	err := os.Mkdir(dir, 0700)

	// we are not going to be stopping on ErrExists errors
	if errors.Is(err, os.ErrExist) {
		err = nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to make directory for acme secrets: %w", err)
	}

	return autocert.DirCache(dir), nil
}
