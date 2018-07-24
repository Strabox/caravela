package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
	"time"
)

func exitFromCaravela(c *cli.Context) {
	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaTimeoutIP(c.GlobalString("ip"), 30*time.Second)

	err := caravelaClient.Exit()
	if err != nil {
		fmt.Printf("Problem exiting the system: %s\n", err)
	}
}
