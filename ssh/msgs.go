package ssh

import (
	"fmt"
)

// directTCPIP request - See RFC4254 7.2 TCP/IP Forwarding Channels
// https://tools.ietf.org/html/rfc4254#page-18
type directTCPIP struct {
	Addr           string
	Rport          uint32
	OriginatorAddr string
	OriginatorPort uint32
}

func (f *directTCPIP) To() string {
	return fmt.Sprintf("%s:%d", f.Addr, f.Rport)
}

// tcpIPForward request - See RFC4254 7.2 TCP/IP Forwarding Channels
// https://tools.ietf.org/html/rfc4254#page-18
type tcpIPForward struct {
	// We have no real use for this address - its supposed to allow the client to specify where
	// the ssh daemon should listen (When GatewayPorts are set to yes), but we dont do any actual listening
	Addr string

	// We use this port do determinane what kind of traffic we should pass along
	// i.e. if the port was 22 - we pass ssh traffic, 80 for http and so on...
	Rport uint32
}

// https://tools.ietf.org/html/rfc4254#section-6.5
type execCommand struct {
	Command string
}
