package anyone

import (
	"os"

	"github.com/fasmide/remotemoe/publickey"
	"golang.org/x/crypto/ssh"
)

func init() {
	enabled := os.Getenv("REMOTEMOE_PUBLICKEY_ANYONE")
	if enabled != "yes" {
		return
	}

	publickey.RegisterSource(&Anyone{})
}

type Anyone struct{}

func (a *Anyone) Authorize(_ string, _ ssh.PublicKey) (bool, error) {
	return true, nil
}
