package services

import (
	"os"
	"strings"
)

// Services holds a map of servicename -> []ports
var Services map[string][]int

// Ports maps port numbers into service names
var Ports map[int]string

// Hostname is the name used when printing out msgs and such
var Hostname string

func init() {
	Services = map[string][]int{
		"http":  {80, 81, 3000, 8000, 8080},
		"https": {443, 3443, 4443, 8443},
		"ssh":   {22, 2222},
	}

	Ports = make(map[int]string)
	for s, ports := range Services {
		for _, p := range ports {
			Ports[p] = s
		}
	}

	h, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	// it is a good idea to keep every host lowercase
	h = strings.ToLower(h)

	Hostname = h
}
