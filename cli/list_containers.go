package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func listContainer(c *cli.Context) {
	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaIP(c.String("ip"))

	containersList, err := caravelaClient.ListContainers()
	if err != nil {
		fmt.Printf("Error with request: %s\n", err)
		os.Exit(1)
	}

	var columnSize = 30
	presentTableLine([]string{
		"CONTAINER ID",
		"IMAGE",
		"STATUS",
		"PORTS"}, columnSize)

	for _, containerStatus := range containersList.ContainersStatus {
		presentTableLine([]string{
			containerStatus.ID,
			containerStatus.ImageKey,
			containerStatus.Status,
			fmt.Sprintf("%v", containerStatus.PortMappings)},
			columnSize)
	}
}

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
