package routertwo

import "fmt"

// Intermediate is able to json parse either Hosts or NamedRoutes from json files
type Intermediate struct {
	Host       *Host       `json:"host,omitempty"`
	NamedRoute *NamedRoute `json:"namedroute,omitempty"`
}

// Wake wakes up a newly parsed Host or NamedRoute
// Named routes needs to know the current router
func (i *Intermediate) Wake(r *Router) (Routable, error) {
	if i.Host != nil {
		return i.Host, nil
	}
	if i.NamedRoute != nil {
		i.NamedRoute.router = r
		return i.NamedRoute, nil
	}

	return nil, fmt.Errorf("invalid json parsed")
}
