package http

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/fasmide/remotemoe/services"
)

type Rewrite struct {
	From Direction

	// upstream
	Scheme string
	Port   string

	// owner is compared against when removing rewrites
	owner string
}

type Direction struct {
	Scheme string
	Host   string
	Port   string
}

type Owner interface {
	FQDN() string
}

func (d *Direction) FromURL(u *url.URL) error {
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	d.Scheme = u.Scheme
	d.Host = host
	d.Port = port

	return nil
}

var (
	lock    sync.RWMutex
	matches map[Direction]Rewrite
	index   map[string][]Rewrite
)

func init() {
	matches = make(map[Direction]Rewrite)
	index = make(map[string][]Rewrite)
}

// Add adds a rewrite to this owners rewrites
func Add(owner Owner, r Rewrite) error {
	// ensure owner is set to the current user
	r.owner = owner.FQDN()

	lock.Lock()
	defer lock.Unlock()

	// add this match or fail
	_, exists := matches[r.From]
	if exists {
		return fmt.Errorf("%+v already exists", r)
	}
	matches[r.From] = r

	// index this new direction for later
	if s, exists := index[owner.FQDN()]; exists {
		index[owner.FQDN()] = append(s, r)
		return nil
	}
	index[owner.FQDN()] = []Rewrite{r}

	return nil
}

func List(owner Owner) []Rewrite {
	lock.RLock()

	match, _ := index[owner.FQDN()]

	lock.RUnlock()

	return match
}

func Remove(owner Owner, d Direction) error {
	lock.Lock()
	defer lock.Unlock()

	m, exists := matches[d]
	if !exists {
		return fmt.Errorf("match %s does not exist", d)
	}

	if m.owner != owner.FQDN() {
		return fmt.Errorf("match %s does not belong to the current user", d)
	}

	// remove directon from matches
	delete(matches, d)

	// remove from index
	s := index[owner.FQDN()]

	// if this is the last entry in the index - remove the index entirely
	if len(s) == 1 {
		delete(index, d.Host)
		return nil
	}

	// search for the direction and chop it off the slice
	for i, r := range s {
		if r.From == d {
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
			break
		}
	}

	return nil
}

func RemoveAll(owner Owner) int {
	lock.Lock()
	defer lock.Unlock()

	all, exists := index[owner.FQDN()]
	if !exists {
		return 0
	}

	delete(index, owner.FQDN())

	for _, r := range all {
		delete(matches, r.From)
	}

	return len(all)
}

func director(r *http.Request) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	localAddr := r.Context().Value(localAddr("localaddr")).(string)
	_, dstPort, _ := net.SplitHostPort(localAddr)

	r.URL.Host = fmt.Sprintf("%s:%s", host, dstPort)

	dPort, _ := strconv.Atoi(dstPort)

	// services.Ports should map 80 into http, 443 into https and so on
	r.URL.Scheme = services.Ports[dPort]

	direction := Direction{}
	err = direction.FromURL(r.URL)
	if err != nil {
		log.Printf("http director: could not determinane direction from request url: %s", err)
		return
	}

	// rewrite direction if a Match exists
	lock.RLock()
	defer lock.RUnlock()
	if re, exists := matches[direction]; exists {
		// change scheme and "host + port"
		r.URL.Scheme = re.Scheme
		r.URL.Host = re.From.Host + ":" + re.Port
	}
}
