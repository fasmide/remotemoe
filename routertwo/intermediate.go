package routertwo

import "fmt"

type Intermediate struct {
	Host       *Host       `json:"host,omitempty"`
	NamedRoute *NamedRoute `json:"namedroute,omitempty"`
}

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
