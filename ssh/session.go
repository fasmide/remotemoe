package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
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
	services map[uint32]struct{}

	// router specifies where we publish the session
	router *router.Router
}

// Handle takes care of a Sessions lifetime
func (s *Session) Handle() {

	// initialize msgs channel
	// we are doing a buffered channel, as a slutty way of not blocking `-N` connections
	// as no terminal is available, we will just buffer them and
	// get on with our lives ... time will tell if this is a good idea :)
	s.msgs = make(chan string, 15)

	// initialize services map
	s.services = make(map[uint32]struct{})

	// if a connection havnt done anything usefull within a minute, throw them away
	s.idleTimeout = time.AfterFunc(IdleTimeout, s.Timeout)

	// take over existing routes
	replaced := s.router.Replace(s)
	if replaced {
		warning := color.New(color.BgYellow, color.FgBlack, color.Bold).Sprint("warn")
		s.msgs <- fmt.Sprintf("%s: this session replaced another session with the same publickey", warning)
	}

	// The incoming Request channel must be serviced.
	go s.handleRequests()

	// block here until the end of time
	s.handleChannels()

	// s.router.Remove will remove this session only if it is the currently active one
	s.router.Remove(s)

	// No reason to keep the timer active
	s.DisableTimeout()
}

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

		if channelRequest.ChannelType() == "session" {
			err := s.acceptSession(channelRequest)
			if err != nil {
				logger.Printf("unable to accept session: %s", err)
			}
			continue
		}

	}
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
			s.services[forwardInfo.Rport] = struct{}{}

			// disable idle timeout now that the connection is actually usefull
			s.DisableTimeout()

			req.Reply(true, nil)
			continue
		}

		logger.Printf("%s: unknown request-type: %s", s.clearConn.RemoteAddr(), req.Type)
		req.Reply(false, nil)

	}

}

// acceptSession starts a new user terminal for the end user
func (s *Session) acceptSession(session ssh.NewChannel) error {
	channel, requests, err := session.Accept()
	if err != nil {
		return fmt.Errorf("unable to accept channel: %w", err)
	}

	// reply "success" to shell and pty-req's
	go func(in <-chan *ssh.Request) {
		for req := range in {
			req.Reply(req.Type == "shell" || req.Type == "pty-req", nil)
		}
	}(requests)

	// setup this sessions terminal
	term := terminal.NewTerminal(channel, "$ ")

	// read commands off the terminal and put them into commands channel
	commands := make(chan string)
	go func() {
		for {
			line, err := term.ReadLine()
			if err != nil {
				close(commands)
				break
			}
			commands <- line
		}
	}()

	fmt.Fprintf(term, "New to remotemoe? - try 'firsttime' or 'help' and start exploring!\r\n")

	go func() {
		defer channel.Close()
		for {
			select {
			case cmd, ok := <-commands:
				if !ok {
					return
				}
				s.handleCommand(cmd, term)
			case msg, ok := <-s.msgs:
				if !ok {
					return
				}
				fmt.Fprintf(term, "%s\r\n", msg)
			}
			s.PokeTimeout()
		}
	}()

	return nil
}

func (s *Session) handleCommand(c string, output io.Writer) {
	bold := color.New(color.Bold)

	switch c {
	case "":
		// nothing
	case "coffie":
		fmt.Fprint(output, "Sure! - have some coffie\r\n")
	case "help":
		bold := color.New(color.Bold)
		bold.Fprint(output, "Commands:")
		fmt.Fprint(output, "\r\n\r\n")

		fmt.Fprint(output, "  services    info about services currently active\r\n")

		fmt.Fprintf(output, "\r\n%s\r\n\r\n", bold.Sprint("Ways of keeping an ssh connection open:"))
		fmt.Fprint(output, "  autossh     using autossh\r\n")
		fmt.Fprint(output, "  unitfile    using a systemd unit\r\n")
		fmt.Fprint(output, "  bashloop    using a simple bash loop\r\n")

		fmt.Fprintf(output, "\r\n%s\r\n\r\n", bold.Sprint("Help topics:"))
		fmt.Fprint(output, "  firstime    first time users of remotemoe and ssh tunneling\r\n")
		fmt.Fprint(output, "  portforward intro to ssh forward ports with `-R`\r\n")

		fmt.Fprint(output, "\r\n")
	case "autossh":
		fmt.Fprintf(output,
			"# autossh template based on ports %s\r\n",
			bold.Sprint(joinDigits(s.serviceKeys())),
		)
		fmt.Fprint(output, "autossh -M 0 -f \\\r\n")
		fmt.Fprint(output, "  -o \"ExitOnForwardFailure yes\" \\\r\n")
		fmt.Fprint(output, "  -o \"ServerAliveInterval 30\" \\\r\n")
		fmt.Fprint(output, "  -o \"ServerAliveCountMax 3\" \\\r\n")

		for p := range s.services {
			fmt.Fprintf(output, "  -R %d:localhost:%d \\\r\n", p, p)
		}

		fmt.Fprintf(output, "  %s -N\r\n", services.Hostname)
		fmt.Fprint(output, "\r\n")
		fmt.Fprint(output, "# for this to work, autossh needs access to the same keys and known_hosts as you had.\r\n")
		fmt.Fprint(output, "# if debugging is needed, remove the `-f` parameter which will keep autossh in the foreground.\r\n")
		fmt.Fprint(output, "\r\n")
	case "unitfile":
		fmt.Fprint(output, "FIXME: Here be unit file\r\n")
	case "bashloop":
		fmt.Fprint(output, "FIXME: Here be bash loop\r\n")
	case "firsttime":
		fmt.Fprintf(output, "%s\r\n", bold.Sprintf("remotemoe"))
		fmt.Fprint(output, "remotemoe allows users to access services that are otherwise inaccessible from the internet.\r\n")
		fmt.Fprint(output, "Just like ngrok or argo tunnels, a device or service connects to remotemoe which in turn muxes\r\n")
		fmt.Fprint(output, "requests back from the internet. \r\n\r\n")

		fmt.Fprintf(output, "%s\r\n", bold.Sprintf("Basic example:"))
		fmt.Fprint(output, "Access the command line and a webservice of a remotely deployed Raspberry Pi:\r\n\r")

		fmt.Fprint(output, firstTimeDiagram)

		fmt.Fprint(output, "\r\n\r\n")
		fmt.Fprint(output, "From the Raspberry pi, connect using `-R` parameters which tells ssh to forward ports.")
		fmt.Fprint(output, "\r\n\r\n")
		fmt.Fprintf(output, "  ssh -R80:localhost:80 -R22:localhost:22 %s\r\n\r\n", services.Hostname)
		fmt.Fprint(output, "That's it, the Raspberry Pi's webservice and ssh daemon are now accessible from the internet\r\n")
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "For information on how to access the services, have a look at the %s command\r\n", bold.Sprintf("services"))
	case "services":

		// Write a few sentences about currently forwarded ports...
		if len(s.services) == 0 {
			fmt.Fprintf(output, "You have %s forwarded ports, have a look in the ssh manual: %s.\r\n", bold.Sprint("zero"), bold.Sprint("man ssh"))
			fmt.Fprintf(output, "You will be looking for the %s parameter.\r\n", bold.Sprint("-R"))
		} else {
			fmt.Fprintf(output,
				"Based on currently forwarded ports %s, your services will be available at:\r\n",
				bold.Sprint(joinDigits(s.serviceKeys())),
			)
		}

		// HTTP services
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "%s (%s)", bold.Sprint("HTTP"), joinDigits(services.Services["http"]))
		fmt.Fprint(output, "\r\n")

		help := true
		for _, p := range services.Services["http"] {
			if _, exists := s.services[uint32(p)]; !exists {
				continue
			}

			// do not display further help about http ports
			help = false

			// port 80 being the default http port - omit the :port format
			if p == 80 {
				fmt.Fprintf(output, "http://%s/\r\n", s.FQDN())
				continue
			}

			fmt.Fprintf(output, "http://%s:%d/\r\n", s.FQDN(), p)
		}

		if help {
			fmt.Fprintf(output, "No HTTP services found, add some by appending `-R80:localhost:80` when connecting.\r\n")
		}

		// HTTPS services
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "%s (%s)", bold.Sprint("HTTPS"), joinDigits(services.Services["https"]))
		fmt.Fprint(output, "\r\n")

		help = true
		for _, p := range services.Services["https"] {
			if _, exists := s.services[uint32(p)]; !exists {
				continue
			}

			// do not display further help about https ports
			help = false

			// port 443 being the default http port - omit the :port format
			if p == 443 {
				fmt.Fprintf(output, "https://%s/\r\n", s.FQDN())
				continue
			}

			fmt.Fprintf(output, "https://%s:%d/\r\n", s.FQDN(), p)
		}

		if help {
			fmt.Fprintf(output, "No HTTPS services found, add some by appending `-R443:localhost:443` when connecting.\r\n")
		}

		// SSH services
		fmt.Fprint(output, "\r\n")
		fmt.Fprintf(output, "%s (%s)", bold.Sprint("SSH"), joinDigits(services.Services["ssh"]))
		fmt.Fprint(output, "\r\n")

		help = true
		for _, p := range services.Services["ssh"] {
			if _, exists := s.services[uint32(p)]; !exists {
				continue
			}

			// do not display further help about ssh ports
			help = false

			// port 22 being the default ssh port - omit the -p<port> format
			if p == 22 {
				fmt.Fprintf(output, "ssh -J %s %s\r\n", services.Hostname, s.FQDN())
				continue
			}

			fmt.Fprintf(output, "ssh -p%d -J %s:%d %s\r\n", p, services.Hostname, p, s.FQDN())
		}

		if help {
			fmt.Fprintf(output, "No SSH services found, add some by appending `-R22:localhost:22` when connecting.\r\n")
		}

	default:
		fmt.Fprintf(output, "%s: command not found\r\n", c)
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
	conn, err := s.router.DialContext(context.Background(), "tcp", forwardInfo.To())
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
	_, isActive := s.services[uint32(p)]
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
	warning := color.New(color.BgYellow, color.FgBlack, color.Bold).Sprint("warn")
	s.msgs <- fmt.Sprintf("%s: this session will be closed, another session just opened with the same publickey, bye!", warning)

	// FIXME: figure out a proper way of flushing msgs to the end user
	time.Sleep(500 * time.Millisecond)

	s.secureConn.Close()
}

func (s *Session) serviceKeys() []int {
	keys := make([]int, 0, len(s.services))
	for v := range s.services {
		keys = append(keys, int(v))
	}
	sort.Sort(sort.IntSlice(keys))
	return keys
}

func joinDigits(ds []int) string {
	b := &strings.Builder{}
	for i, v := range ds {
		if i == 0 {
			fmt.Fprintf(b, "%d", v)
			continue
		}
		if i == len(ds)-1 {
			fmt.Fprintf(b, " and %d", v)
			continue
		}
		fmt.Fprintf(b, ", %d", v)
	}
	return b.String()
}
