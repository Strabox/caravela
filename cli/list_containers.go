package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
	"strings"
)

func listContainer(c *cli.Context) {
	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaIP(c.GlobalString("ip"))

	containersStatus, err := caravelaClient.ListContainers()
	if err != nil {
		fatalPrintf("Error with request: %s\n", err)
	}

	var columnSize = 30
	presentTableLine([]string{
		"CONTAINER ID",
		"IMAGE",
		"STATUS",
		"PORTS",
		"NAME"}, columnSize)

	for _, containerStatus := range containersStatus {

		presentPortMappings := ""
		for i, portMap := range containerStatus.PortMappings {
			if i != 0 {
				presentPortMappings += ", "
			}
			presentPortMappings += fmt.Sprintf("%s:%d->%d/tcp", containerStatus.SupplierIP, portMap.HostPort,
				portMap.ContainerPort)
		}

		presentTableLine([]string{
			containerStatus.ContainerID,
			containerStatus.ImageKey,
			containerStatus.Status,
			presentPortMappings,
			containerStatus.Name},
			columnSize)
	}
}

// presentTableLine formats information
func presentTableLine(information []string, columnSize int) {
	for _, info := range information {
		toPrintInfo := info
		if len(toPrintInfo) > columnSize {
			toPrintInfo = info[:columnSize]
		} else {
			toPrintInfo += strings.Repeat(" ", columnSize-len(toPrintInfo))
		}
		fmt.Print(toPrintInfo)
	}
	fmt.Println()
}
