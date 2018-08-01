package cli

import (
	"github.com/urfave/cli"
	"net"
)

func join(c *cli.Context) {
	if c.NArg() < 2 {
		fatalPrintln("Please provide the host IP address and the join node IP address")
	}

	hostIP := c.Args().Get(0)
	if net.ParseIP(hostIP) == nil {
		fatalPrintln("Please provide a valid host IP address")
	}

	joinIP := c.Args().Get(1)
	if net.ParseIP(joinIP) == nil {
		fatalPrintln("Please provide a valid join IP address")
	}

	if err := initNode(hostIP, true, joinIP); err != nil {
		fatalPrintf("Error: %s\n", err)
	}
}
