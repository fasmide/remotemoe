package ssh

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// ChannelConn embedds ssh.Channel and is used to fulfill the net.Conn interface
type ChannelConn struct {
	ssh.Channel
}

// LocalAddr fulfills net.Conn
func (c *ChannelConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr fulfills net.Conn
func (c *ChannelConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline fulfills net.Conn
func (c *ChannelConn) SetDeadline(_ time.Time) error {
	return nil
}

// SetReadDeadline fulfills net.Conn
func (c *ChannelConn) SetReadDeadline(_ time.Time) error {
	return nil
}

// SetWriteDeadline fulfills net.Conn
func (c *ChannelConn) SetWriteDeadline(_ time.Time) error {
	return nil
}
