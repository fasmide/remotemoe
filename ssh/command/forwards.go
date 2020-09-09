package command

import (
	"bytes"
	"fmt"

	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

const forwardDiagram = `
            +------------------------> remotemoe ignores bind_address
            |        +---------------> specifies what kind of service is being forwarded
            |        |    +----------> destination host
            |        |    |     +----> destination port
            +        +    +     +
  -R [bind_address:]port:host:hostport
`

func Forwards(d Forwarding) *cobra.Command {
	help := &bytes.Buffer{}
	fmt.Fprintf(help, "Forwards:\n")
	fmt.Fprint(help, "  List all currently forwarded ports\n\n")
	fmt.Fprint(help, "  These are ports that was forwarded by the client by using ssh's `-R` parameter.\n")
	fmt.Fprint(help, "  remotemoe uses the ports and hostnames like this:\n")
	fmt.Fprint(help, forwardDiagram)
	fmt.Fprint(help, "  \n")
	fmt.Fprint(help, "  Incoming port forwards are mapped directly to service mux'es, with the following rules:\n\n")
	fmt.Fprintf(help, "  Ports %s will be accessible with %s\n", joinDigits(services.Services["http"]), "HTTP")
	fmt.Fprintf(help, "  Ports %s will be accessible with %s\n", joinDigits(services.Services["https"]), "HTTPs")
	fmt.Fprintf(help, "  Ports %s will be accessible with %s\n", joinDigits(services.Services["ssh"]), "ssh")
	fmt.Fprintf(help, "  Other ports can be accessed though ssh by using `ssh -L` or `ssh -W`\n")

	return &cobra.Command{
		Use:   "forwards",
		Short: "List all currently forwarded ports",
		Long:  help.String(),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Forwarded ports:")

			f := d.Forwards()
			for p := range f {
				s, exists := services.Ports[int(p)]
				if !exists {
					s = "other"
				}

				cmd.Printf("%d/%s\n", p, s)
			}

			if len(f) == 0 {
				cmd.Printf("None\n")
			}
		},
	}
}
