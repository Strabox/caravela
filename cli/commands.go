package cli

import "github.com/urfave/cli"

var (
	commands = []cli.Command{
		{
			Name:      "join",
			ShortName: "j",
			Usage:     "Join a caravela instance",
			Category:  "Caravela instance management",
			Action:    join,
		},
		{
			Name:      "create",
			ShortName: "c",
			Usage:     "Create a caravela instance",
			Category:  "Caravela instance management",
			Action:    create,
		},
		{
			Name:      "run",
			ShortName: "r",
			Usage:     "Launch a container in the Caravela instance",
			Category:  "Node management",
			Action:    run,
			Flags: []cli.Flag{
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
