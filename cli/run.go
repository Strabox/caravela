package cli

import (
	"fmt"
	"github.com/strabox/caravela/api/client"
	"github.com/urfave/cli"
	"os"
)

func run(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Println("Please provide at least a container image to launch")
		os.Exit(1)
	}

	var containerArgs []string = nil

	if c.NArg() > 1 {
		containerArgs = make([]string, c.NArg()-1)
		for i := 1; i < c.NArg(); i++ {
			containerArgs[i-1] = c.Args().Get(i)
		}
	}

	caravelaClient := client.NewCaravela(c.String("ip"))

	err := caravelaClient.Run(c.Args().Get(0), containerArgs, int(c.Uint("cpus")), int(c.Uint("ram")))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
