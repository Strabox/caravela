package cli

import (
	"context"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
	"time"
)

func exitFromCaravela(c *cli.Context) {
	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaTimeoutIP(c.GlobalString("ip"), 30*time.Second)

	err := caravelaClient.Shutdown(context.Background())
	if err != nil {
		fatalPrintf("Problem exiting the system: %s\n", err)
	}
}
