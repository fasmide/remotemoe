package publickey

import (
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

type Source interface {
	Authorize(string, ssh.PublicKey) (bool, error)
}

var sources []Source

func init() {
	sources = make([]Source, 0)
}

func RegisterSource(s Source) {
	sources = append(sources, s)
	log.Printf("%T publickey plugin registered", s)
}

func Authorize(u string, k ssh.PublicKey) (bool, error) {
	var wg sync.WaitGroup
	for _, s := range sources {
		wg.Add(1)
		go func(s Source) {
			wg.Done()
		}(s)
	}

	wg.Wait()
}
