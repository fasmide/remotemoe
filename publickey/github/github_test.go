package github

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"golang.org/x/crypto/ssh"
)

var testKeys = []string{
	"keys_test/id_ecdsa.pub",
	"keys_test/id_ed25519.pub",
	"keys_test/id_rsa.pub",
}

type MockGithub struct {
}

func (m *MockGithub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("request for keys: %s", r.URL)
	for _, v := range testKeys {
		k, err := os.ReadFile(v)
		if err != nil {
			panic(err)
		}

		w.Write(k)
	}
}

type MockBrokenGithub struct {
	http.ServeMux
}

func NewMockBrokenGithub() *MockBrokenGithub {
	m := &MockBrokenGithub{}
	m.HandleFunc("/brokenKeys/", m.brokenKeys)
	m.HandleFunc("/longresponse/", m.longResponse)
	return m
}

func (m *MockBrokenGithub) longResponse(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request for long response")
	for i := 0; i < 100000000; i++ {
		w.Write([]byte("h"))
	}
}

func (m *MockBrokenGithub) brokenKeys(w http.ResponseWriter, r *http.Request) {
	log.Printf("request for broken keys: %s", r.URL)
	for _, v := range testKeys {
		k, err := os.ReadFile(v)
		if err != nil {
			panic(err)
		}

		w.Write(k)
		w.Write([]byte("\nlarhblarh broken not at all a ssh key\n"))
	}
}

func TestTooLongResponse(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("could not listen: %s", err)
	}

	go http.Serve(l, NewMockBrokenGithub())

	e := fmt.Sprintf("http://%s/longresponse/", l.Addr().String())
	g := Github{Endpoint: e}

	_, err = g.PublicKeys("someone")
	if err != nil {
		return
	}
	t.Fatalf("PublicKeys did return keys on broken response")
}

func TestBrokenGithub(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("could not listen: %s", err)
	}

	go http.Serve(l, NewMockBrokenGithub())

	e := fmt.Sprintf("http://%s/brokenKeys/", l.Addr().String())
	g := Github{Endpoint: e}

	_, err = g.PublicKeys("someone")
	if err != nil {
		return
	}
	t.Fatalf("PublicKeys did return keys on broken response")
}

func TestUnavailableGithub(t *testing.T) {
	g := Github{Endpoint: "http://127.127.127.127:1337"}

	_, err := g.PublicKeys("someone-that-does-not-exist")
	if err == nil {
		t.Fatalf("non existing github succeded")
		return
	}
}

func TestGithub(t *testing.T) {
	k, err := os.ReadFile("keys_test/id_rsa.pub")
	if err != nil {
		t.Fatalf("could not read test public key: %s", err)
	}

	key, _, _, _, err := ssh.ParseAuthorizedKey(k)
	if err != nil {
		t.Fatalf("could not parse public key: %s", err)
	}

	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("could not listen: %s", err)
	}

	go http.Serve(l, &MockGithub{})

	e := fmt.Sprintf("http://%s/", l.Addr().String())
	g := Github{Endpoint: e}

	keys, err := g.PublicKeys("someone")

	for _, k := range keys {
		if bytes.Equal(key.Marshal(), k.Marshal()) {
			return
		}
	}

	t.Fail()
}

func TestGithubNegative(t *testing.T) {
	k, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("unable to generate key: %s", err)
	}
	key, err := ssh.NewPublicKey(k)
	if err != nil {
		t.Fatalf("unable to make ssh public key: %s", err)
	}

	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatalf("could not listen: %s", err)
	}

	go http.Serve(l, &MockGithub{})

	e := fmt.Sprintf("http://%s/", l.Addr().String())
	g := Github{Endpoint: e}

	keys, err := g.PublicKeys("someone")

	for _, k := range keys {
		if bytes.Equal(key.Marshal(), k.Marshal()) {
			t.Fatalf("non existing key found")
		}
	}

}
