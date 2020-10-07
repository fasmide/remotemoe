package rewrite

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fasmide/remotemoe/http"
	"github.com/fasmide/remotemoe/router"
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

// Add allows the user to add a new match
func Add(r router.Routable) *cobra.Command {
	c := &cobra.Command{
		Use:   "add",
		Short: "Add a new match",
		RunE: func(cmd *cobra.Command, args []string) error {
			// both urls must have a host that belongs to the current ssh session
			match, err := url.Parse(args[0])
			if err != nil {
				return fmt.Errorf("unable to parse match url %s: %w", args[0], err)
			}

			err = validateURL(match, r)
			if err != nil {
				return fmt.Errorf("unable to validate match: %w", err)
			}

			// if the match host contains a portnumber, make sure the match port and scheme makes sense
			if port := match.Port(); port != "" {
				p, err := strconv.Atoi(port)
				if err != nil {
					return fmt.Errorf("unable to parse port form match url: %w", err)
				}

				if pScheme, found := services.Ports[p]; found {
					if pScheme != match.Scheme {
						return fmt.Errorf("port %d of match url will never encounter %s traffic, only %s", p, match.Scheme, pScheme)
					}
				}
			}

			dest, err := url.Parse(args[1])
			if err != nil {
				return fmt.Errorf("unable to parse destination url %s: %w", args[1], err)
			}

			err = validateURL(dest, r)
			if err != nil {
				return fmt.Errorf("unable to validate destination url: %w", err)
			}

			var m http.Direction
			err = m.FromURL(match)
			if err != nil {
				return fmt.Errorf("unable to parse match: %w", err)
			}

			var d http.Direction
			err = d.FromURL(dest)
			if err != nil {
				return fmt.Errorf("unable to parse destination: %w", err)
			}

			rewrite := http.Rewrite{From: m, To: d}
			err = http.Add(rewrite)
			if err != nil {
				return fmt.Errorf("unable to add match to http router: %w", err)
			}

			cmd.Printf("%s to %s ... no problem\n", match, dest)

			return nil
		},
		Args: cobra.ExactArgs(2),
	}

	return c

}

func validateURL(u *url.URL, creator router.Routable) error {
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
	if namedRoute, ok := r.(*router.NamedRoute); ok {
		if namedRoute.Owner != creator.FQDN() {
			return fmt.Errorf("this session does not own %s", host)
		}
	} else {
		// if this is not a named route, host should match the current session's FQDN
		if host != creator.FQDN() {
			return fmt.Errorf("this session cannot add matches for other sessions hostnames")
		}
	}

	err := validateScheme(u.Scheme)
	if err != nil {
		return fmt.Errorf("unknown scheme: %w", err)
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
