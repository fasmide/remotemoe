package routertwo

import (
	"encoding/json"
	"fmt"
)

// Intermediate is able to json parse either Hosts or NamedRoutes from json files
type Intermediate struct {
	Host       *Host       `json:"host,omitempty"`
	NamedRoute *NamedRoute `json:"namedroute,omitempty"`

	Metadata map[string]json.RawMessage `json:"metadata"`
}

func NewIntermediate(e *Entry) (*Intermediate, error) {
	// we need to convert Marshaler into "[]byte"
	md := make(map[string]json.RawMessage)
	for n, v := range e.Metadata {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal %s: %w", n, err)
		}

		md[n] = json.RawMessage(b)
	}
	i := &Intermediate{Metadata: md}

	switch r := e.Routable.(type) {
	case *Host:
		i.Host = r
	case *NamedRoute:
		i.NamedRoute = r
	default:
		return nil, fmt.Errorf("unknown routable type: %T", e.Routable)
	}

	return i, nil
}

// Wake wakes up a newly parsed Host or NamedRoute
// Named routes needs to know the current router
func (i *Intermediate) Wake(r *Router) (Routable, map[string]interface{}, error) {
	md := make(map[string]interface{})
	for n, v := range i.Metadata {
		dec, exists := decoders[n]
		if !exists {
			return nil, nil, fmt.Errorf("unable to decode \"%s\" without a decoder", n)
		}

		marshaler, err := dec.Decode(v)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to decode json: %w", err)
		}
		md[n] = marshaler
	}

	if i.Host != nil {
		return i.Host, md, nil
	}
	if i.NamedRoute != nil {
		i.NamedRoute.router = r
		return i.NamedRoute, md, nil
	}

	return nil, nil, fmt.Errorf("invalid json parsed")
}
