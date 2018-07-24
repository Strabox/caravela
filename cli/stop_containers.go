package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
	"os"
)

func stopContainers(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Println("Please provide at least a container ID to be stopped")
		os.Exit(1)
	}

	containersIDs := make([]string, c.NArg())
	for i := 0; i < c.NArg(); i++ {
		containersIDs[i] = c.Args().Get(i)
	}

	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaIP(c.GlobalString("ip"))

	err := caravelaClient.StopContainers(containersIDs)
	if err != nil {
		fmt.Printf("Problem stopping the containers: %s\n", err)
	}
}
