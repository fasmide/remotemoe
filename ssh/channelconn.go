package ssh

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// ChannelConn embedds a ssh.Channel and implements dummy methods to fulfill the net.Conn interface
type ChannelConn struct {
	ssh.Channel
}

// LocalAddr is required by net.Conn
func (c *ChannelConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr is required by net.Conn
func (c *ChannelConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline is required by net.Conn
func (c *ChannelConn) SetDeadline(_ time.Time) error {
	return nil
}

// SetReadDeadline is required by net.Conn
func (c *ChannelConn) SetReadDeadline(_ time.Time) error {
	return nil
}

// SetWriteDeadline is required by net.Conn
func (c *ChannelConn) SetWriteDeadline(_ time.Time) error {
	return nil
}
