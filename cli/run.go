package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
)

const DefaultNumOfCPUs = 1
const DefaultAmountOfRAM = 256

func run(c *cli.Context) {

	if c.NArg() < 1 {
		log.Fatal("Please provide at least a container image key")
	}

	var containerArgs []string = nil

	if c.NArg() > 1 {
		containerArgs = make([]string, c.NArg()-1)
		for i := 1; i < c.NArg(); i++ {
			containerArgs[i-1] = c.Args().Get(i)
		}
	}

	caravelaClient := client.NewCaravela()

	caravelaClient.Run(c.Args().Get(0), containerArgs, int(c.Uint("cpus")), int(c.Uint("ram")))
}
