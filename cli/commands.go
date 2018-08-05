package cli

import "github.com/urfave/cli"

// List of commands available to the CLI end users.
var (
	commands = []cli.Command{
		{
			Name:      "join",
			ShortName: "j",
			Usage:     "Join a caravela instance",
			Category:  "Caravela system management",
			Before:    printBanner,
			Action:    join,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "hostIP, hip",
					Usage: "Host's IP address",
					Value: defaultHostIP,
				},
			},
		},
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "Create a caravela instance",
			Category:  "Caravela system management",
			Before:    printBanner,
			Action:    create,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "hostIP, hip",
					Usage: "Host's IP address",
					Value: defaultHostIP,
				},
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Configuration's file path",
					Value: defaultConfigurationFile,
				},
			},
		},
		{
			Name:      "run",
			ShortName: "r",
			Usage:     "Launch a container in the Caravela instance",
			Category:  "User's containers management",
			Action:    runContainers,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Usage: "Name for the container",
					Value: defaultContainerName,
				},
				cli.StringSliceFlag{
					Name:  "portMap, p",
					Usage: "Define a port mapping for a container, HostPort:ContainerPort",
					Value: &cli.StringSlice{}, // No predefined port mapping
				},
				cli.StringFlag{
					Name:  "cpuPower, cp",
					Usage: "Power/Class of the CPU necessary for the container",
					Value: defaultCPUPower,
				},
				cli.UintFlag{
					Name:  "cpus, c",
					Usage: "Maximum number of CPUs/Cores that the container need",
					Value: defaultCPUs,
				},
				cli.UintFlag{
					Name:  "ram, r",
					Usage: "Maximum amount of RAM (in Megabytes) that container can use",
					Value: defaultRAM,
				},
			},
		},
		{
			Name:     "container",
			Aliases:  []string{"c"},
			Usage:    "Options for managing user's containers",
			Category: "User's containers management",
			Before:   printBanner,
			Subcommands: []cli.Command{
				{
					Name:   "ps",
					Usage:  "List the user's containers in the system",
					Action: listContainer,
				},
				{
					Name:   "stop",
					Usage:  "Stop a set of containers",
					Action: stopContainers,
				},
			},
		},
		{
			Name:      "exit",
			ShortName: "e",
			Usage:     "Shutdown from the CARAVELA instance, makes the node leave",
			Category:  "Caravela system management",
			Before:    printBanner,
			Action:    exitFromCaravela,
		},
	}
)
