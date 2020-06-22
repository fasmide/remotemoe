package main

import (
	"flag"
	"log"
	"net"

	"github.com/fasmide/remotemoe/http"
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/ssh"
)

var sshPort = flag.Int("sshport", 0, "ssh listen port")
var httpPort = flag.Int("httpport", 0, "http listen port")

func main() {
	flag.Parse()

	router := router.New()

	proxy := &http.HttpProxy{Router: router}
	proxy.Initialize()

	httpListener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: *httpPort})
	if err != nil {
		log.Fatalf("cannot listen for ssh connections: %s", err)
	}
	HTTPServer := http.New()
	HTTPServer.Handler = proxy
	go func() {
		log.Printf("http server failed: %s", HTTPServer.Listen(httpListener))
		// we dont care if this http server fails
	}()

	sshConfig, err := ssh.DefaultConfig()
	if err != nil {
		log.Fatalf("cannot get default ssh config: %s", err)
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: *sshPort})
	if err != nil {
		log.Fatalf("cannot listen for ssh connections: %s", err)
	}

	log.Print("ssh listening on ", listener.Addr())

	sshServer := ssh.Server{Config: sshConfig, Router: router}
	sshServer.Listen(listener)
}
