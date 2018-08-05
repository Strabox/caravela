package cli

import (
	"context"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
)

func stopContainers(c *cli.Context) {
	if c.NArg() < 1 {
		fatalPrintln("Please provide at least a container ID to be stopped")
	}

	containersIDs := make([]string, c.NArg())
	for i := 0; i < c.NArg(); i++ {
		containersIDs[i] = c.Args().Get(i)
	}

	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaIP(c.GlobalString("ip"))

	err := caravelaClient.StopContainers(context.Background(), containersIDs)
	if err != nil {
		fatalPrintf("Problem stopping the containers: %s\n", err)
	}
}
