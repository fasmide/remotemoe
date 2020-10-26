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
}

type Direction struct {
	Scheme string
	Host   string
	Port   string
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

func Add(r Rewrite) error {
	lock.Lock()
	defer lock.Unlock()

	// add this match or fail
	_, exists := matches[r.From]
	if exists {
		return fmt.Errorf("%+v already exists", r)
	}
	matches[r.From] = r

	// index this new direction for later
	if s, exists := index[r.From.Host]; exists {
		index[r.From.Host] = append(s, r)
		return nil
	}
	index[r.From.Host] = []Rewrite{r}

	return nil
}

func List(hosts ...string) []Rewrite {
	lock.RLock()

	result := []Rewrite{}
	for _, host := range hosts {
		if match, exists := index[host]; exists {
			result = append(result, match...)
		}
	}

	lock.RUnlock()
	return result
}

func Remove(d Direction) {
	lock.Lock()
	defer lock.Unlock()

	// remove directon from matches
	delete(matches, d)

	// remove from index
	s := index[d.Host]

	// if this is the last entry in the index - remove the index entirely
	if len(s) == 1 {
		delete(index, d.Host)
		return
	}

	// search for the direction and chop it off the slice
	for i, r := range s {
		if r.From == d {
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
			break
		}
	}

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
