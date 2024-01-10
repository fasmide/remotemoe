package github

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/crypto/ssh"
)

var validGithubUsername = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`).MatchString

func PublicKey(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	user := c.User()

	if !validGithubUsername(user) {
		return nil, fmt.Errorf("%s is not an allowed github username", user)
	}

	u, err := url.JoinPath("https://github.com/", user+".keys")
	if err != nil {
		log.Printf("unable to lookup keys for %s: %s", user, err)
		return nil, fmt.Errorf("unable to lookup keys for %s: %w", user, err)
	}

	resp, err := http.Get(u)
	if err != nil {
		log.Printf("unable to fetch keys for %s: %s", user, err)
		return nil, fmt.Errorf("unable to fetch keys for %s: %w", user, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("unable to fetch keys for %s: unexpected status code %d", user, resp.StatusCode)
		return nil, fmt.Errorf("unable to fetch keys for %s: unexpected status code %d", user, resp.StatusCode)
	}

	// the usual id_rsa is some 400 bytes, lets make it 512 to add room for some
	// comments and lets multiply it with 20, should be enough room for 20 rsa keys which is
	// the largest ones ssh supports
	lr := io.LimitReader(resp.Body, 512*20)

}
