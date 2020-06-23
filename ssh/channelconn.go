package ssh

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type ChannelConn struct {
	ssh.Channel
}

func (c *ChannelConn) LocalAddr() net.Addr {
	return nil
}
func (c *ChannelConn) RemoteAddr() net.Addr {
	return nil
}

func (c *ChannelConn) SetDeadline(_ time.Time) error {
	return nil
}

func (c *ChannelConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (c *ChannelConn) SetWriteDeadline(_ time.Time) error {
	return nil
}
