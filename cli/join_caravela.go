package cli

import (
	"fmt"
	"github.com/urfave/cli"
	"net"
	"os"
)

func join(c *cli.Context) {
	if c.NArg() < 2 {
		fmt.Println("Please provide the host IP address and the join node IP address")
		os.Exit(1)
	}

	hostIP := c.Args().Get(0)
	if net.ParseIP(hostIP) == nil {
		fmt.Println("Please provide a valid host IP address")
		os.Exit(1)
	}

	joinIP := c.Args().Get(1)
	if net.ParseIP(joinIP) == nil {
		fmt.Println("Please provide a valid join IP address")
		os.Exit(1)
	}

	if err := initNode(hostIP, true, joinIP); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
