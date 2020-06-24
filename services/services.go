package services

// Services holds a map of servicename -> []ports
var Services map[string][]int

// Ports maps port numbers into service names
var Ports map[int]string

func init() {
	Services = map[string][]int{
		"HTTP":  {80, 81, 3000, 8000, 8080},
		"HTTPS": {443, 3443, 4443, 8443},
		"SSH":   {22, 2222},
	}

	Ports = make(map[int]string)
	for s, ports := range Services {
		for _, p := range ports {
			Ports[p] = s
		}
	}
}
