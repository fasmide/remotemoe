package main

import (
	"flag"
	"log"
	"net"

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

	server := http.NewServer()
	server.Handler = proxy

	for _, port := range services.Services["HTTP"] {
		l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
		if err != nil {
			// listen errors are usually not a big deal, maybe the user just
			// dont have permissions on privileged-ports but we are still able to work
			// on higher portnumbers
			log.Printf("cannot accept HTTP on %d: %s", port, err)
			continue
		}

		log.Printf("accepting %s on %d", "HTTP", port)
		go server.Serve(l)
	}

	sshConfig, err := ssh.DefaultConfig()
	if err != nil {
		log.Fatalf("cannot get default ssh config: %s", err)
	}

	sshServer := ssh.Server{Config: sshConfig, Router: router}

	for _, port := range services.Services["SSH"] {
		l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
		if err != nil {
			log.Printf("cannot accept SSH on %d: %s", port, err)
			continue
		}

		log.Printf("accepting %s on %d", "SSH", port)
		go sshServer.Serve(l)
	}

	select {}
}
