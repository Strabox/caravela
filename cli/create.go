package cli

import (
	"fmt"
	"github.com/urfave/cli"
	"net"
	"os"
)

func create(c *cli.Context) {
	if c.NArg() < 1 {
		fmt.Println("Please provide the host IP address")
		os.Exit(1)
	}

	hostIP := c.Args().Get(0)
	if net.ParseIP(hostIP) == nil {
		fmt.Println("Invalid host IP address")
		os.Exit(1)
	}

	initNode(hostIP, false, "")
}
