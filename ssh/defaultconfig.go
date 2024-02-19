package ssh

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"os"

	"github.com/fasmide/hostkeys"
	"github.com/fasmide/remotemoe/publickey"
	"golang.org/x/crypto/ssh"

	_ "github.com/fasmide/remotemoe/publickey/anyone"
	_ "github.com/fasmide/remotemoe/publickey/github"
)

// RawPrivateKey could be set with ldflags on build time
var RawPrivateKey string

const noPublicKeyBanner = `                            __                              
.----.-----.--------.-----.|  |_.-----.--------.-----.-----.
|   _|  -__|        |  _  ||   _|  -__|        |  _  |  -__|
|__| |_____|__|__|__|_____||____|_____|__|__|__|_____|_____|

You somehow forgot to present a public key when trying to authenticate.
Please continue by creating one:

	ssh-keygen

remotemoe accepts any key - see ya!`

type PublicKeyCB func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error)

// DefaultConfig generates a default ssh.ServerConfig
func DefaultConfig() (*ssh.ServerConfig, error) {
	config := &ssh.ServerConfig{
		BannerCallback: func(_ ssh.ConnMetadata) string {
			return "the remote.moe service now requires authentication\nread all about it:https://github.com/fasmide/remotemoe/discussions/14\n"
		},

		// try to take advantage of AES-NI, by moving chachapoly last of preferred ciphers
		// 	* Well that didnt work - it seems the official ssh client really likes chacha20,
		//	so if we really want AES-NI it seems we need to drop support for chacha20
		Config: ssh.Config{
			Ciphers: []string{
				"aes128-gcm@openssh.com",
				"aes128-ctr",
				"aes192-ctr",
				"aes256-ctr",
				// "chacha20-poly1305@openssh.com",
			},
		},
		MaxAuthTries: 1,
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			allow, err := publickey.Authorize(c.User(), pubKey)
			if err != nil {
				return nil, fmt.Errorf("unable to authorize: %w", err)
			}

			if !allow {
				return nil, fmt.Errorf("denied")
			}

			return &ssh.Permissions{
				// Record the public key used for authentication.
				Extensions: map[string]string{
					"pubkey-fp":  ssh.FingerprintSHA256(pubKey),
					"pubkey-ish": fingerprintIsh(pubKey),
					"pubkey":     string(ssh.MarshalAuthorizedKey(pubKey)),
				},
			}, nil
		},
		// We will use the keyboard interactive auth method as a way of telling the user that
		// he needs to create a public key and use that instead - we should not get here if the user already has
		// a working key and presented that in the first place
		KeyboardInteractiveCallback: func(conn ssh.ConnMetadata, client ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
			_, err := client(conn.User(), noPublicKeyBanner, []string{""}, []bool{false})
			if err != nil {
				return nil, fmt.Errorf("error doing keyboard interactive challenge: %w", err)
			}

			return nil, fmt.Errorf("user did not public key")
		},
	}

	m := hostkeys.Manager{
		Directory: os.Getenv("CONFIGURATION_DIRECTORY"),
	}

	err := m.Manage(config)
	if err != nil {
		return nil, fmt.Errorf("problem managing host keys: %w", err)
	}

	return config, nil
}

// fingerprintIsh is named after commitish as the plan originally was
// to just use the last 8 or so bits of the sha256 and just live with the
// risk of collisions but now we are doing the whole sum of the public key, but (lowercased) base32 encoded
// as base64 is not very friendly for use in host names
func fingerprintIsh(pubKey ssh.PublicKey) string {
	sha256sum := sha256.Sum256(pubKey.Marshal())
	enc := base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)
	return enc.EncodeToString(sha256sum[:])
}
