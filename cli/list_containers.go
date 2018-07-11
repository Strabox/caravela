package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
)

func listContainer(c *cli.Context) {
	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaIP(c.String("ip"))

	containersList, err := caravelaClient.ListContainers()
	if err != nil {
		fmt.Printf("Problem exiting the system: %s\n", err)
	}

	fmt.Println("ID                      STATUS")
	for _, containerStatus := range containersList.ContainersStatus {
		fmt.Printf("%s                      %s\n", containerStatus.ID, "TODO")
	}
}
