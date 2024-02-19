package publickey

import (
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"
)

type Source interface {
	Authorize(string, ssh.PublicKey) (bool, error)
}

// at some point it might be interresting to support
// multiple sources at once - for now, lets go with
// only one source at a time
var source Source

func RegisterSource(s Source) {
	if source != nil {
		log.Fatalf("publickey source: cannot use %T, %T already selected", s, source)
	}
	source = s
	log.Printf("publickey source: using %T for authentication", s)
}

func Authorize(u string, k ssh.PublicKey) (bool, error) {
	if source == nil {
		log.Printf("publickey: no authentication source chosen")
		return false, fmt.Errorf("no public key source")
	}
	return source.Authorize(u, k)
}
