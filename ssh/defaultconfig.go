package ssh

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

// RawPrivateKey could be set with ldflags on build time
var RawPrivateKey string

// DefaultConfig generates a default ssh.ServerConfig
func DefaultConfig() (*ssh.ServerConfig, error) {
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},

		// Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{
				// Record the public key used for authentication.
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					"pubkey":    string(ssh.MarshalAuthorizedKey(pubKey)),
				},
			}, nil
		},
	}

	signer, err := signer()
	if err != nil {
		return nil, err
	}
	config.AddHostKey(signer)

	return config, nil
}

// signer returns a ssh.Signer from RawPrivateKey or by looking for id_rsa files
func signer() (ssh.Signer, error) {

	// if no private key shipped with this binary try to read
	// id_rsa from the working directory
	if RawPrivateKey == "" {
		privateBytes, err := ioutil.ReadFile("id_rsa")
		if err != nil {
			return nil, fmt.Errorf("Failed to load private key: %s", err)
		}

		signer, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse private key: %s", err)
		}

		return signer, nil
	}

	// if this binary ships with a private key - use that
	private, err := ssh.ParsePrivateKey([]byte(RawPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse embedded private key: %s", err)
	}

	signer, ok := private.(ssh.Signer)
	if !ok {
		return nil, fmt.Errorf("cannot cast %T to ssh.Signer", private)
	}

	return signer, nil

}
