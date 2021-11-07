package rewrite

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fasmide/remotemoe/http"
	"github.com/fasmide/remotemoe/routertwo"
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

// Add allows the user to add a new match
func Add(r routertwo.Routable, router *routertwo.Router) *cobra.Command {
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a new match",

		RunE: func(cmd *cobra.Command, args []string) error {
			match, err := url.Parse(args[0])
			if err != nil {
				return fmt.Errorf("unable to parse match url %s: %w", args[0], err)
			}

			// we have some extra requirements for these urls, make sure they are fulfiled
			err = validateURL(match, r, router)
			if err != nil {
				return fmt.Errorf("unable to validate match: %w", err)
			}

			var m http.Direction
			err = m.FromURL(match)
			if err != nil {
				return fmt.Errorf("unable to parse match: %w", err)
			}

			// port and scheme from flags
			flags := cmd.LocalFlags()

			// scheme
			scheme, err := flags.GetString("to-scheme")
			if err != nil {
				return fmt.Errorf("scheme flag error: %w", err)
			}

			// incase scheme is empty i.e. not set, use scheme from url
			if scheme == "" {
				scheme = m.Scheme
			}

			// port
			port, err := flags.GetString("to-port")
			if err != nil {
				return fmt.Errorf("port flag error: %w", err)
			}

			// incase port is empty i.e. not set, use port from url
			if port == "" {
				port = m.Port
			}

			rewrite := http.Rewrite{
				From:   m,
				Scheme: scheme,
				Port:   port,
			}

			err = http.Add(r, rewrite)
			if err != nil {
				return fmt.Errorf("unable to add match to http router: %w", err)
			}

			cmd.Printf("%s to %s/%s ... no problem\n", match, scheme, port)

			return nil
		},
		Args: cobra.ExactArgs(1),
	}

	c.Flags().StringP("to-scheme", "s", "", "scheme to be used upstream")
	c.Flags().StringP("to-port", "p", "", "port to be used upstream")

	return c

}

func validateURL(u *url.URL, creator routertwo.Routable, router *routertwo.Router) error {
	// urls cannot be relative paths i.e. they must have a host
	if u.Host == "" {
		return fmt.Errorf("no host provided in url: %s", u.String())
	}

	// u.Host may be in the form of "host:port" - split off port if its there
	host := u.Host
	if u.Port() != "" {
		host = strings.SplitN(host, ":", 2)[0]
	}

	// host must be available in the router
	r, found := router.Find(host)
	if !found {
		return fmt.Errorf("host \"%s\" not found, add with `host add %s`", host, host)
	}

	// if r is a namedRoute, it must be owned by the current routable
	if namedRoute, ok := r.(*routertwo.NamedRoute); ok {
		if namedRoute.Owner != creator.FQDN() {
			return fmt.Errorf("this session does not own %s", host)
		}
	} else {
		// if this is not a named route, host should match the current session's FQDN
		if host != creator.FQDN() {
			return fmt.Errorf("this session cannot add matches for other sessions hostnames")
		}
	}

	// if the match host contains a portnumber, make sure the match port and scheme makes sense
	if port := u.Port(); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("unable to parse port form match url: %w", err)
		}

		if pScheme, found := services.Ports[p]; found {
			if pScheme != u.Scheme {
				return fmt.Errorf("port %d of match url will never encounter %s traffic, only %s", p, u.Scheme, pScheme)
			}
		}
	}

	return validateScheme(u.Scheme)
}

func validateScheme(s string) error {
	if s == "http" {
		return nil
	}
	if s == "https" {
		return nil
	}

	return fmt.Errorf("scheme is neigher http or https")

}
