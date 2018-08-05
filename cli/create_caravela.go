package cli

import (
	"github.com/urfave/cli"
	"net"
)

func create(c *cli.Context) {
	hostIP := c.String("hostIP")
	if hostIP == "" {
		hostIP = getOutboundIP()
	}
	if net.ParseIP(hostIP) == nil {
		fatalPrintf("Invalid host IP address: %s\n", hostIP)
	}

	if err := initNode(hostIP, c.String("config"), false, ""); err != nil {
		fatalPrintf("Problem: %s\n", err)
	}
}
