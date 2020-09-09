package command

import (
	"github.com/fasmide/remotemoe/services"
	"github.com/spf13/cobra"
)

const firstTimeDiagram = `
        raspberry pi
       +------------------------------------+
       |$ ssh -R22:localhost:22 \           |
       |      -R80:localhost:80 remotemoe   |
       +---------------+^^+-----------------+
                       |**|
   corporate firewall  |**|
  +--------------------|**|-----------------------+
                       |**|
                       |**| http and ssh traffic are
                       |**| wrapped inside ssh tunnel
                       |**|
  +--------------------|**|-----------------------+
   internet            |**|
                       |**|
        remotemoe      |**|
       +---------------v--v-----------------+
       |maps services such as http, https   |
       |and ssh to ssh tunnels.             |
       |                                    |
       +-------^----------------------^-----+
               *                      *
               * http traffic         * ssh traffic
   internet    *                      *
  +------------*----------------------*-----------+
               *                      *
   browser     *           ssh client *
  +------------+---------+ +----------+-----------+
  |$ curl key.remotemoe  | |$ ssh -J remotemoe key|
  |                      | |                      |
  +----------------------+ +----------------------+
`

func Firsttime() *cobra.Command {
	return &cobra.Command{
		Use:   "firsttime",
		Short: "Explains what remotemoe is all about",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("%s\n", "remotemoe")
			cmd.Print("remotemoe allows users to access services that are otherwise inaccessible from the internet.\n")
			cmd.Print("Just like ngrok or argo tunnels, a device or service connects to remotemoe which in turn muxes ")
			cmd.Print("requests back from the internet. \n\n")

			cmd.Printf("%s\n", "Basic example:")
			cmd.Print("Access the command line and a webservice of a remotely deployed Raspberry Pi:\n")

			cmd.Print(firstTimeDiagram)

			cmd.Print("\n\n")
			cmd.Print("From the Raspberry pi, connect using `-R` parameters which tells ssh to forward ports.")
			cmd.Print("\n\n")
			cmd.Printf("  ssh -R80:localhost:80 -R22:localhost:22 %s\n\n", services.Hostname)
			cmd.Print("That's it, the Raspberry Pi's webservice and ssh daemon are now accessible from the internet\n")
		},
	}
}
