package cli

import (
	"github.com/urfave/cli"
	"net"
)

func create(c *cli.Context) {
	if c.NArg() < 1 {
		fatalPrintln("Please provide the host IP address")
	}

	hostIP := c.Args().Get(0)
	if net.ParseIP(hostIP) == nil {
		fatalPrintf("Invalid host IP address: %s\n", hostIP)
	}

	if err := initNode(hostIP, false, ""); err != nil {
		fatalPrintf("Problem: %s\n", err)
	}
}
