package router

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/asdine/storm/v3"
)

// Routable describes requirements to be routable
type Routable interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)

	// A routable should be able to identify it self
	FQDN() string

	// Replaced is used to indicate to a ssh session that it's routes
	// have been replaced with another ssh session
	Replaced()
}

var lock sync.RWMutex
var endpoints map[string]Routable

var db *storm.DB

// ErrNotFound errors are returned when searching for something
// that does not exist
type ErrNotFound error

// Initialize loads previously stored namedroutes and sets everything up
func Initialize() error {
	endpoints = make(map[string]Routable)

	// open database
	var err error
	db, err = storm.Open(databasePath())
	if err != nil {
		return fmt.Errorf("router: unable to open database: %s", err)
	}

	// fetch previously stored named routes
	var namedRoutes []*NamedRoute
	err = db.All(&namedRoutes)
	if err != nil {
		return fmt.Errorf("router: database error: %s", err)
	}

	// restore all named routes
	for _, namedRoute := range namedRoutes {
		endpoints[namedRoute.FQDN()] = namedRoute
	}

	log.Printf("router: restored %d routes, using %s", len(namedRoutes), databasePath())

	return nil
}

func databasePath() string {
	dirs := strings.FieldsFunc(
		os.Getenv("STATE_DIRECTORY"),
		func(r rune) bool {
			return r == ':'
		},
	)

	// dirs[0] is the acme directory
	if len(dirs) > 1 {
		return path.Join(dirs[1], "router.db")
	}

	return "router.db"
}

// Add is used to add a named route
func Add(n *NamedRoute) error {
	lock.Lock()
	defer lock.Unlock()

	existing, exists := endpoints[n.FQDN()]
	if exists {
		// if the found route is a *NamedRoute that happens to belong
		// to the same owner as the provided one, return without an error
		if existingNamedRoute, ok := existing.(*NamedRoute); ok {
			if existingNamedRoute.Owner == n.Owner {
				return nil
			}
		}

		return fmt.Errorf("router: %s is already active", n.FQDN())
	}

	err := db.Save(n)
	if err != nil {
		log.Printf("router: cannot add %s: %s", n.Name, err)
		return errors.New("could not store name, try again later")
	}

	endpoints[n.FQDN()] = n

	return nil
}

// RemoveName removes a named router if
// * The route exists
// * The route is a *NamedRoute
// * The *NamedRoute's owner, is the one trying to remove it
func RemoveName(s string, from Routable) error {
	lock.Lock()
	defer lock.Unlock()

	toRemove, exists := endpoints[s]
	if !exists {
		return fmt.Errorf("%s does not exist", s)
	}

	// we must ensure the route that was found, is a namedroute
	namedRouteToRemove, ok := toRemove.(*NamedRoute)
	if !ok {
		return fmt.Errorf("%s is not a named route", s)
	}

	// the Routable must own this namedRoute
	if namedRouteToRemove.Owner != from.FQDN() {
		return fmt.Errorf("%s is not your route to remove", s)
	}

	err := db.DeleteStruct(toRemove)
	if err != nil {
		log.Printf("router: could not remove %s: %s", s, err)
		return fmt.Errorf("%s cannot be removed, try again later", s)
	}

	// if we have come this far, the route exists, it is a named route and does belong to the user trying to remove it
	delete(endpoints, toRemove.FQDN())

	return nil
}

// RemoveAll clears every hostname from a routable
func RemoveAll(s Routable) ([]*NamedRoute, error) {
	lock.Lock()
	defer lock.Unlock()

	var toRemove []*NamedRoute

	err := db.Find("Owner", s.FQDN(), &toRemove)
	if errors.Is(err, storm.ErrNotFound) {
		return toRemove, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot find names to remove: %s", err)
	}

	for _, nr := range toRemove {
		err := db.DeleteStruct(nr)
		if err != nil {
			log.Printf("router: could not remove %s: %s", nr.FQDN(), err)
			return nil, fmt.Errorf("unable to remove all names")
		}
		delete(endpoints, nr.FQDN())
	}

	return toRemove, nil
}

// Names returns a list of NamedRoutes
func Names(r Routable) ([]NamedRoute, error) {
	var result []NamedRoute

	err := db.Find("Owner", r.FQDN(), &result)
	if errors.Is(err, storm.ErrNotFound) {
		err = nil
	}

	if err != nil {
		err = fmt.Errorf("router: unable to fetch names: %w", err)
		log.Printf(err.Error())
		return nil, err
	}

	return result, nil
}

// Replace replaces a route, if another route was found
// the old route will have its .Replaced function called
// and this method will return true - if this was a new
// route, it will return false
func Replace(d Routable) bool {
	n := d.FQDN()

	lock.Lock()
	oldRoute, exists := endpoints[n]
	if exists {
		go oldRoute.Replaced()
	}

	endpoints[n] = d
	lock.Unlock()

	return exists
}

// Remove removes a route by name and Routable
func Remove(d Routable) {
	n := d.FQDN()

	lock.Lock()
	defer lock.Unlock()

	endpoint, ok := endpoints[n]
	if !ok {
		return
	}

	if d == endpoint {
		delete(endpoints, n)
	}

}

// Find fetches a route
func Find(n string) (Routable, bool) {
	lock.RLock()
	d, ok := endpoints[n]
	lock.RUnlock()
	return d, ok
}

// Exists allows the acme/autocert to figure out if it should make certificate requests
func Exists(_ context.Context, s string) error {
	lock.RLock()
	_, exists := endpoints[s]
	lock.RUnlock()

	if !exists {
		return fmt.Errorf("%s does not exist", s).(ErrNotFound)
	}

	return nil
}

// DialContext is used by stuff that what to dial something up
func DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("router: could not split host from port: %w", err)
	}

	lock.RLock()
	d, exists := endpoints[host]
	lock.RUnlock()

	if !exists {
		return nil, fmt.Errorf("router: %s not found", host).(ErrNotFound)
	}

	return d.DialContext(ctx, network, address)
}
