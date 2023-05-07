package routertwo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// ErrNotFound errors are returned when searching for something
// that does not exist
var ErrNotFound = errors.New("not found")

// Routable describes requirements to be routable
type Routable interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)

	// A routable should be able to identify it self
	FQDN() string

	// Replaced is used to indicate to a ssh session that it's routes
	// have been replaced with another ssh session
	Replaced()
}

// Router - takes care of Routable and Namedroutes
type Router struct {
	sync.RWMutex

	dbPath string

	editLock sync.Mutex
	a        *map[string]Routable
	b        *map[string]Routable
	active   *map[string]Routable

	nameIndex map[string][]*NamedRoute
}

// NewRouter initializes a new Router with a given path
func NewRouter(dbPath string) (*Router, error) {
	// make room for a and b lists
	a := make(map[string]Routable)
	b := make(map[string]Routable)

	r := &Router{
		dbPath:    dbPath,
		a:         &a,
		b:         &b,
		nameIndex: make(map[string][]*NamedRoute),
	}

	r.active = r.a

	err := filepath.WalkDir(dbPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fd, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("unable to open %s: %w", p, err)
		}
		defer fd.Close()

		dec := json.NewDecoder(fd)
		var i Intermediate
		err = dec.Decode(&i)
		if err != nil {
			return fmt.Errorf("unable to decode json (%s): %w", p, err)
		}

		routable, err := i.Wake(r)
		if err != nil {
			return fmt.Errorf("json format error (%s): %w", p, err)
		}

		a[routable.FQDN()] = routable
		b[routable.FQDN()] = routable

		nroute, ok := routable.(*NamedRoute)
		if ok {
			r.index(nroute)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to bring router database up: %w", err)
	}

	return r, nil
}

// DialContext is used by stuff that what to dial something up
func (r *Router) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("router: could not split host from port: %w", err)
	}

	r.RLock()
	d, exists := (*r.active)[host]
	r.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s not found", ErrNotFound, host)
	}

	return d.DialContext(ctx, network, address)
}

// Online should only be used by peers, e.g. ssh clients which proved
// by authentication that they do infact have the private key for their FQDN
func (r *Router) Online(rtbl Routable) (bool, error) {
	next, old := r.begin()
	defer r.finish()

	var host *Host
	var replaced bool
	oldRoute, exists := (*next)[rtbl.FQDN()]
	if exists { // route exists
		var ok bool
		host, ok = oldRoute.(*Host)
		if ok { // route is host
			go host.Replaced()
			if host.Routable != nil {
				replaced = true
			}

			host = &Host{
				Routable: rtbl,
				Name:     rtbl.FQDN(),
				LastSeen: time.Now(),
				Created:  host.Created,
			}
		} else { // if route was not host - just replace
			host = &Host{
				Routable: rtbl,
				Name:     rtbl.FQDN(),
				LastSeen: time.Now(),
				Created:  time.Now(),
			}
		}
	} else { // we havnt seen this one before
		host = &Host{
			Routable: rtbl,
			Name:     rtbl.FQDN(),
			LastSeen: time.Now(),
			Created:  time.Now(),
		}
	}

	// store this host on disk
	i := &Intermediate{Host: host}
	err := r.store(rtbl.FQDN(), i)
	if err != nil {
		return false, fmt.Errorf("unable to store host: %w", err)
	}

	// do the exchange
	(*next)[rtbl.FQDN()] = host

	r.exchange(next)

	(*old)[rtbl.FQDN()] = host

	return replaced, nil
}

// Offline removes the routable from a host
func (r *Router) Offline(d Routable) {
	next, old := r.begin()
	defer r.finish()

	// we should be able to find this routable
	routable, ok := (*r.active)[d.FQDN()]
	if !ok {
		return
	}

	// and the routable should be a host type
	host, ok := routable.(*Host)
	if !ok {
		return
	}

	// and the host, should contain this actual routable
	if d != host.Routable {
		return
	}

	// change last seen and update record
	host.LastSeen = time.Now()

	i := &Intermediate{Host: host}
	err := r.store(host.FQDN(), i)
	if err != nil {
		// we need to continue even if we encounter this error
		log.Printf("router: unable to update host as it went offline: %s", err)
	}

	// we have to create a new Host and have the old one garbage collected
	host = &Host{
		Routable: nil,
		Name:     host.Name,
		LastSeen: host.LastSeen,
		Created:  host.Created,
	}

	// do the exchange
	(*next)[host.Name] = host

	r.exchange(next)

	(*old)[host.Name] = host

}

// AddName adds a *NamedRoute to the router
func (r *Router) AddName(n *NamedRoute) error {
	next, old := r.begin()
	defer r.finish()

	// existing routes are handled differently
	existing, exists := (*next)[n.FQDN()]
	if exists {
		if existingNamedRoute, ok := existing.(*NamedRoute); ok {
			if existingNamedRoute.Owner == n.Owner {
				return nil
			}
		}

		return fmt.Errorf("%s is occupied", n.FQDN())
	}

	// make sure this name is able to use us
	n.router = r

	// handle new routes
	i := &Intermediate{NamedRoute: n}
	err := r.store(n.FQDN(), i)
	if err != nil {
		return fmt.Errorf("unable to store route: %w", err)
	}

	r.index(n)

	(*next)[n.FQDN()] = n

	r.exchange(next)

	(*old)[n.FQDN()] = n

	return nil
}

// RemoveName removes a named router if
// * The route exists
// * The route is a *NamedRoute
// * The *NamedRoute's owner, is the one trying to remove it
func (r *Router) RemoveName(s string, from Routable) error {
	next, old := r.begin()
	defer r.finish()

	toRemove, exists := (*next)[s]
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

	// remove from disk
	err := r.unlink(s)
	if err != nil {
		return fmt.Errorf("fs error: %w", err)
	}

	r.reduceIndex(from.FQDN(), namedRouteToRemove)

	delete((*next), s)

	r.exchange(next)

	delete((*old), s)

	return nil
}

// RemoveNames removes all names from a Routable
func (r *Router) RemoveNames(from Routable) ([]*NamedRoute, error) {
	next, old := r.begin()
	defer r.finish()

	list, exists := r.nameIndex[from.FQDN()]
	if !exists {
		return make([]*NamedRoute, 0), nil
	}

	// remove namedroutes until we hit the first error
	var successes int
	var err error
	for i, n := range list {
		err = r.unlink(n.FQDN())
		if err != nil {
			break
		}

		successes = i

		r.reduceIndex(from.FQDN(), n)

		delete((*next), n.FQDN())
	}

	r.exchange(next)

	// remove the same from the old list
	for i := 0; i <= successes; i++ {
		delete((*old), list[i].FQDN())
	}
	return list[:successes+1], err

}

// Names returns a list of NamedRoutes
func (r *Router) Names(rtbl Routable) ([]NamedRoute, error) {
	r.RLock()
	defer r.RUnlock()

	n, exists := r.nameIndex[rtbl.FQDN()]
	if !exists {
		return make([]NamedRoute, 0), nil
	}

	names := make([]NamedRoute, 0, len(n))
	for _, nr := range n {
		names = append(names, *nr)
	}

	return names, nil
}

// Find fetches a route
func (r *Router) Find(n string) (Routable, bool) {
	r.RLock()
	d, exists := (*r.active)[n]
	r.RUnlock()

	return d, exists
}

// Exists returns an error if a given hostname does not exist
func (r *Router) Exists(_ context.Context, s string) error {
	r.RLock()
	_, exists := (*r.active)[s]
	r.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s not found", ErrNotFound, s)
	}

	return nil
}

func (r *Router) index(value *NamedRoute) {
	_, exists := r.nameIndex[value.Owner]
	if exists {
		r.nameIndex[value.Owner] = append(r.nameIndex[value.Owner], value)
	} else {
		r.nameIndex[value.Owner] = []*NamedRoute{value}
	}
}

func (r *Router) reduceIndex(key string, value *NamedRoute) {
	i, exists := r.nameIndex[key]
	if !exists {
		return
	}

	// now, we need to find the value
	var idx int
	var found bool
	for i, n := range i {
		if n == value {
			idx = i
			found = true
			break
		}
	}

	if !found {
		return
	}

	if len(i) == 1 {
		delete(r.nameIndex, key)
		return
	}

	// we make a kind of copy of the index excluding the index in question
	ret := make([]*NamedRoute, 0)
	ret = append(ret, i[:idx]...)
	r.nameIndex[key] = append(ret, i[idx+1:]...)
}

func (r *Router) begin() (*map[string]Routable, *map[string]Routable) {
	r.editLock.Lock()
	if r.active == r.a {
		return r.b, r.a
	}

	return r.a, r.b
}

func (r *Router) exchange(m *map[string]Routable) {
	r.Lock()
	r.active = m
	r.Unlock()
}

func (r *Router) finish() {
	r.editLock.Unlock()
}

func (r *Router) store(n string, i *Intermediate) error {
	p := path.Join(r.dbPath, fmt.Sprint(n, ".json"))

	fd, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("unable to store data: %w", err)
	}

	defer fd.Close()

	enc := json.NewEncoder(fd)
	err = enc.Encode(i)
	if err != nil {
		return fmt.Errorf("unable to encode data: %w", err)
	}

	return nil
}

func (r *Router) unlink(n string) error {
	p := path.Join(r.dbPath, fmt.Sprint(n, ".json"))

	err := os.Remove(p)

	// If this file did not exist - its okay
	if errors.Is(err, syscall.ENOENT) {
		return nil
	}

	return err
}
