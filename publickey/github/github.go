package github

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/crypto/ssh"
)

type Github struct {
	Endpoint string
}

var validGithubUsername = regexp.MustCompile(`^[a-zA-Z0-9\-]+$`).MatchString

// PublicKeys returns public keys from github
func (g *Github) PublicKeys(user string) ([]ssh.PublicKey, error) {
	if !validGithubUsername(user) {
		return nil, fmt.Errorf("%s is not an allowed github username", user)
	}

	u, err := url.JoinPath(g.Endpoint, user+".keys")
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

	keys := make([]ssh.PublicKey, 0)
	scanner := bufio.NewScanner(lr)
	for scanner.Scan() {
		pk, _, _, _, err := ssh.ParseAuthorizedKey(scanner.Bytes())
		if err != nil {
			log.Printf("failed to parse public keys from %s: %s", u, err)
			return nil, fmt.Errorf("failed to parse public keys from %s: %w", u, err)
		}

		keys = append(keys, pk)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("unable to read keys from http response: %s", err)
		return nil, fmt.Errorf("unable to read keys from http response: %w", err)
	}

	return keys, nil

}
