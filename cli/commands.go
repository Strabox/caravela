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
		},
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "Create a caravela instance",
			Category:  "Caravela system management",
			Before:    printBanner,
			Action:    create,
		},
		{
			Name:      "run",
			ShortName: "r",
			Usage:     "Launch a container in the Caravela instance",
			Category:  "User's containers management",
			Action:    runContainers,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "portMap, p",
					Value: &cli.StringSlice{}, // No predefined port mapping
					Usage: "Define a port mapping for a container, HostPort:ContainerPort",
				},
				cli.UintFlag{
					Name:  "cpus, c",
					Value: 0,
					Usage: "Maximum number of CPUs/Cores that the container need",
				},
				cli.UintFlag{
					Name:  "ram, r",
					Value: 0,
					Usage: "Maximum amount of RAM (in Megabytes) that container can use",
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
					Name:   "ls",
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
			Usage:     "Exit from the CARAVELA instance, makes the node leave",
			Category:  "Caravela system management",
			Before:    printBanner,
			Action:    exitFromCaravela,
		},
	}
)
