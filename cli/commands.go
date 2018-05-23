package cli

import "github.com/urfave/cli"

/*
List of all commands available to the user.
*/
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
			Category:  "Caravela node management",
			Action:    run,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "portMap, p",
					Value: &cli.StringSlice{}, // No predefined port mapping
					Usage: "Define a port mapping for a container, HostPort:ContainerPort",
				},
				cli.StringFlag{
					Name:  "ip",
					Value: DefaultCaravelaInstanceIP,
					Usage: "IP of the caravela instance to send the request",
				},
				cli.UintFlag{
					Name:  "cpus, c",
					Value: DefaultNumOfCPUs,
					Usage: "Maximum number of CPUs that the container can use",
				},
				cli.UintFlag{
					Name:  "ram, r",
					Value: DefaultAmountOfRAM,
					Usage: "Maximum amount of RAM (in Megabytes) that container can use",
				},
			},
		},
	}
)
