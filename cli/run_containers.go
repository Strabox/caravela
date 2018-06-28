package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/util"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
)

func runContainers(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Println("Please provide at least a container image to launch")
		os.Exit(1)
	}

	// Validate port mappings provided
	var portMappings = make([]rest.PortMapping, 0)
	fmt.Printf("Port Mappings: %v\n", c.StringSlice("p"))
	for _, portMap := range c.StringSlice("p") {
		var err error
		resultPortMap := rest.PortMapping{}
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
		if !util.IsValidPort(resultPortMap.HostPort) {
			fmt.Printf("Invalid host port number in: %s\n", portMap)
			os.Exit(1)
		}

		if resultPortMap.ContainerPort, err = strconv.Atoi(containerPort); err != nil {
			fmt.Printf("Port should be a positive integer. Error: %s.\n", containerPort)
			os.Exit(1)
		}
		if !util.IsValidPort(resultPortMap.ContainerPort) {
			fmt.Printf("Invalid container port number in: %s\n", portMap)
			os.Exit(1)
		}

		portMappings = append(portMappings, resultPortMap)
	}

	// Obtains all the arguments provided to the container launch
	var containerArgs []string = nil
	if c.NArg() > 1 {
		containerArgs = make([]string, c.NArg()-1)
		for i := 1; i < c.NArg(); i++ {
			fmt.Printf("Aarg %d: %v\n", i, c.Args().Get(i))
			containerArgs[i-1] = c.Args().Get(i)
		}
	}

	// Create a user client of the CARAVELA system
	caravelaClient := client.NewCaravelaIP(c.String("ip"))

	err := caravelaClient.RunContainer(c.Args().Get(0), c.StringSlice("p"), containerArgs,
		int(c.Uint("cpus")), int(c.Uint("ram")))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
