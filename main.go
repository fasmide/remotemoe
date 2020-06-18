package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/ssh"
)

var sshPort = flag.Int("sshport", 0, "ssh listen port")
var httpPort = flag.Int("httpport", 0, "http listen port")

func main() {
	flag.Parse()

	go func() {
		l := fmt.Sprintf(":%d", *httpPort)
		log.Print("http listening on ", l)
		log.Printf("http server failed: %s", http.ListenAndServe(l, nil))
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

	router := router.New()
	sshServer := ssh.Server{Config: sshConfig, Router: router}
	sshServer.Listen(listener)
}
