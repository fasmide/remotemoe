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

	err := router.Initialize()
	if err != nil {
		log.Fatalf("could not initialize router: %s", err)
	}

	proxy := &http.HttpProxy{}
	proxy.Initialize()

	server, err := http.NewServer()
	if err != nil {
		panic(err)
	}

	server.Handler = proxy

	services.Serve("http", server)
	services.ServeTLS("https", server)

	sshConfig, err := ssh.DefaultConfig()
	if err != nil {
		log.Fatalf("cannot get default ssh config: %s", err)
	}

	sshServer := &ssh.Server{Config: sshConfig}

	services.Serve("ssh", sshServer)

	// we shall be dealing with shutting down in the future :)
	select {}
}
