package ssh

import "io"

type Router struct {
	Endpoints map[string]io.ReadWriter
}
