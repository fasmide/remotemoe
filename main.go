package main

import (
	"flag"
	"log"

	"github.com/fasmide/remotemoe/http"
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
	"github.com/fasmide/remotemoe/ssh"
)

func main() {
	flag.Parse()

	router := router.New()

	proxy := &http.HttpProxy{Router: router}
	proxy.Initialize()

	server := http.NewServer(router)
	server.Handler = proxy

	services.Serve("HTTP", server)
	services.ServeTLS("HTTPS", server)

	sshConfig, err := ssh.DefaultConfig()
	if err != nil {
		log.Fatalf("cannot get default ssh config: %s", err)
	}

	sshServer := &ssh.Server{Config: sshConfig, Router: router}

	services.Serve("SSH", sshServer)

	// we shall be dealing with shutting down in the future :)
	select {}
}
