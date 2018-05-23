package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/api/rest"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
)

func run(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Println("Please provide at least a container image to launch")
		os.Exit(1)
	}

	// Validate port mappings provided
	var portMappings = make([]rest.PortMappingJSON, 0)
	fmt.Printf("PortMaps: %v\n", c.StringSlice("p"))
	for _, portMap := range c.StringSlice("p") {
		var err error
		resultPortMap := rest.PortMappingJSON{}
		portMapping := strings.Split(portMap, ":")

		if len(portMapping) != 2 {
			fmt.Printf("Invalid port mapping: %s. Provide: HostPort:ContainerPort\n", portMapping)
			os.Exit(1)
		}

		hostPort := portMapping[0]
		containerPort := portMapping[1]
		if resultPortMap.HostPort, err = strconv.Atoi(hostPort); err != nil {
			fmt.Printf("Port should be a positive integer. Error: %s.\n", hostPort)
			os.Exit(1)
		}
		if resultPortMap.ContainerPort, err = strconv.Atoi(containerPort); err != nil {
			fmt.Printf("Port should be a positive integer. Error: %s.\n", containerPort)
			os.Exit(1)
		}

		portMappings = append(portMappings, resultPortMap)
	}

	// Obtains all the arguments provided to the container launch
	var containerArgs []string = nil
	if c.NArg() > 1 {
		containerArgs = make([]string, c.NArg()-1)
		for i := 1; i < c.NArg(); i++ {
			containerArgs[i-1] = c.Args().Get(i)
		}
	}

	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaIP(c.String("ip"))

	err := caravelaClient.Run(c.Args().Get(0), containerArgs, c.StringSlice("p"),
		int(c.Uint("cpus")), int(c.Uint("ram")))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
