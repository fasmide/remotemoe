package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/fasmide/remotemoe/http"
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
)

// IdleTimeout sets how long a session can be idle before getting disconnected
const IdleTimeout = time.Minute

// Session represents a ongoing SSH connection
type Session struct {
	// Raw socket
	clearConn net.Conn
	// SSH Socket
	secureConn *ssh.ServerConn
	// Incoming channel requests
	channelRequests <-chan ssh.NewChannel
	// Global requests
	requests <-chan *ssh.Request

	idleLock     sync.Mutex
	idleTimeout  *time.Timer
	idleDisabled bool

	// messages to the terminal (i.e. the user)
	msgs chan string

	// services list of forwarded port numbers
	// these are just indicators that the remote sent a tcpip-forward request sometime
	services     map[uint32]struct{}
	servicesLock sync.RWMutex

	// registeOnce is used to register with the router when ever a
	// forward is received ... but only once :)
	registerOnce sync.Once
}

// Handle takes care of a Sessions lifetime
func (s *Session) Handle() {

	// initialize msgs channel
	// we are doing a buffered channel, as a slutty way of not blocking `-N` connections
	// as no terminal is available, we will just buffer them and
	// get on with our lives ... time will tell if this is a good idea :)
	s.msgs = make(chan string, 50)

	// initialize services map
	s.services = make(map[uint32]struct{})

	// if a connection havnt done anything useful within a minute, throw them away
	s.idleTimeout = time.AfterFunc(IdleTimeout, s.Timeout)

	// The incoming Request channel must be serviced.
	go s.handleRequests()

	// block here until the end of time
	s.handleChannels()

	// router.Remove will remove this session only if it is the currently active one
	router.Remove(s)

	// http.RemoveAll will remove all http rules this session may have set up
	http.RemoveAll(s)

	// No reason to keep the timer active
	s.DisableTimeout()
}

// Close closes a ssh session
func (s *Session) Close() error {
	return s.secureConn.Close()
}

// FQDN returns the fully qualified hostname for this session
func (s *Session) FQDN() string {
	return fmt.Sprintf("%s.%s", s.secureConn.Permissions.Extensions["pubkey-ish"], services.Hostname)
}

// Timeout fires when the session has done too much idling
func (s *Session) Timeout() {
	logger.Printf("%s idle for more then %s:, closing", s.clearConn.RemoteAddr(), IdleTimeout)

	err := s.secureConn.Close()
	if err != nil {
		logger.Printf("could not close secureConnection: %s", err)
	}

}

// PokeTimeout postprones the idle timer - unless disabled or already fired
func (s *Session) PokeTimeout() {
	s.idleLock.Lock()
	defer s.idleLock.Unlock()

	if s.idleDisabled {
		return
	}

	if !s.idleTimeout.Stop() {
		// too late - it already fired
		return
	}

	s.idleTimeout.Reset(IdleTimeout)
}

// DisableTimeout disables the idle timeout, used when a connection provides some endpoints
// i.e. requests ports to be forwarded...
func (s *Session) DisableTimeout() {
	s.idleLock.Lock()
	defer s.idleLock.Unlock()

	s.idleTimeout.Stop()
	s.idleDisabled = true
}

func (s *Session) handleChannels() {
	for channelRequest := range s.channelRequests {
		// direct-tcpip forward requests
		if channelRequest.ChannelType() == "direct-tcpip" {
			err := s.acceptForwardRequest(channelRequest)
			if err != nil {
				logger.Printf("unable to accept channel: %s", err)
			}

			continue
		}

		// "shell requests"
		if channelRequest.ChannelType() == "session" {
			c := Console{session: s}
			err := c.Accept(channelRequest)
			if err != nil {
				logger.Printf("unable to accept session: %s", err)
			}

			continue
		}

		logger.Printf("unknown ChannelType from %s: %s", s.secureConn.RemoteAddr(), channelRequest.ChannelType())
	}
}

// Forwards returns a copy of forwarded port numbers
func (s *Session) Forwards() map[uint32]struct{} {
	s.servicesLock.RLock()

	v := make(map[uint32]struct{})
	for p := range s.services {
		v[p] = struct{}{}
	}

	s.servicesLock.RUnlock()

	return v
}

func (s *Session) handleRequests() {

	for req := range s.requests {
		if req.Type == "keepalive@openssh.com" {
			req.Reply(true, nil)
			continue
		}

		if req.Type == "tcpip-forward" {
			forwardInfo := tcpIPForward{}
			err := ssh.Unmarshal(req.Payload, &forwardInfo)

			if err != nil {
				logger.Printf("%s: unable to unmarshal forward information: %s", s.clearConn.RemoteAddr(), err)
				req.Reply(false, nil)
				continue
			}

			// store this port number in services - future Dial's to this session
			// will know if the service is available by looking in there
			s.servicesLock.Lock()
			s.services[forwardInfo.Rport] = struct{}{}
			s.servicesLock.Unlock()

			// disable idle timeout now that the connection is actually useful
			s.DisableTimeout()

			// register with the router - only do this once
			s.registerOnce.Do(func() {
				// take over existing routes
				replaced := router.Replace(s)
				if replaced {
					warning := color.New(color.BgYellow, color.FgBlack, color.Bold)
					warning.EnableColor()
					s.msgs <- fmt.Sprintf("%s: this session replaced another session with the same publickey\n", warning.Sprint("warn"))
				}
			})

			s.informForward(forwardInfo.Rport)

			req.Reply(true, nil)
			continue
		}

		logger.Printf("%s: unknown request-type: %s", s.clearConn.RemoteAddr(), req.Type)
		req.Reply(false, nil)

	}

}

// informForward informs the user that the forward request have been accepted and where its available
func (s *Session) informForward(p uint32) {
	bold := color.New(color.Bold)
	bold.EnableColor()

	// first things first - do we know what to do with this portnumber?
	service, exists := services.Ports[int(p)]
	if !exists {
		s.msgs <- fmt.Sprintf("%s (%d)\nssh -L%d:%s:%d %s\n", bold.Sprintf("other"), p, p, s.FQDN(), p, services.Hostname)
		return
	}

	switch service {
	case "http": // http services
		if p == 80 {
			s.msgs <- fmt.Sprintf("%s (%d)\nhttp://%s/\n", bold.Sprintf("http"), p, s.FQDN())
		} else {
			s.msgs <- fmt.Sprintf("%s (%d)\nhttp://%s:%d/\n", bold.Sprintf("http"), p, s.FQDN(), p)
		}
	case "https": // https services
		if p == 443 {
			s.msgs <- fmt.Sprintf("%s (%d)\nhttps://%s/\n", bold.Sprintf("https"), p, s.FQDN())
		} else {
			s.msgs <- fmt.Sprintf("%s (%d)\nhttps://%s:%d/\n", bold.Sprintf("https"), p, s.FQDN(), p)
		}
	case "ssh": // ssh services
		if p == 22 {
			s.msgs <- fmt.Sprintf("%s (%d)\nssh -J %s %s\n", bold.Sprintf("ssh"), p, services.Hostname, s.FQDN())
		} else {
			s.msgs <- fmt.Sprintf("%s (%d)\nssh -p%d -J %s:%d %s\n", bold.Sprintf("ssh"), p, p, services.Hostname, p, s.FQDN())
		}
	default:
		s.msgs <- fmt.Sprintf("erhm port %d - a certain developer must be ashamed of it self :)", p)
	}
}

// acceptForwardRequest parses information about the request, checks to see if an endpoint
// matching is available in the router and then io.Copy'es everything back and forth
func (s *Session) acceptForwardRequest(fr ssh.NewChannel) error {
	forwardInfo := directTCPIP{}
	err := ssh.Unmarshal(fr.ExtraData(), &forwardInfo)
	if err != nil {
		fr.Reject(ssh.UnknownChannelType, "failed to parse forward information")
		return fmt.Errorf("unable to unmarshal forward information: %w", err)
	}

	// lookup "hostname" in the router, fetch remote and pass data
	conn, err := router.DialContext(context.Background(), "tcp", forwardInfo.To())
	if err != nil {
		err = fmt.Errorf("cannot dial %s: %s", forwardInfo.To(), err)
		fr.Reject(ssh.ConnectionFailed, fmt.Sprintf("cannot make connection: %s", err))
		return err
	}

	// Accept channel from ssh client
	channel, requests, err := fr.Accept()
	if err != nil {
		return fmt.Errorf("could not accept forward channel: %w", err)
	}

	// we should not timeout this client - its talking to another client
	s.DisableTimeout()

	go ssh.DiscardRequests(requests)

	go io.Copy(channel, conn)
	go io.Copy(conn, channel)

	return nil
}

// DialContext tries to dial connections though the ssh session
// FIXME: figure out what to do with the Context
func (s *Session) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("unable to figure out host and port: %w", err)
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("unable to convert port number to int: %w", err)
	}

	// did the client forward this port prior to this request?
	s.servicesLock.RLock()
	_, isActive := s.services[uint32(p)]
	s.servicesLock.RUnlock()

	if !isActive {
		return nil, fmt.Errorf("this client does not provide port %d", p)
	}

	channel, reqs, err := s.secureConn.OpenChannel("forwarded-tcpip", ssh.Marshal(directTCPIP{
		Addr:  "localhost",
		Rport: uint32(p),
	}))

	if err != nil {
		return nil, fmt.Errorf("could not open remote channel: %w", err)
	}

	go ssh.DiscardRequests(reqs)

	cConn := &ChannelConn{Channel: channel}
	return cConn, nil

}

// Replaced is called when another ssh session is replacing this current one
func (s *Session) Replaced() {
	warning := color.New(color.BgYellow, color.FgBlack, color.Bold)
	warning.EnableColor()

	s.msgs <- fmt.Sprintf("%s: this session will be closed, another session just opened with the same publickey, bye!", warning.Sprint("warn"))

	// FIXME: figure out a proper way of flushing msgs to the end user
	time.Sleep(500 * time.Millisecond)

	s.secureConn.Close()
}
