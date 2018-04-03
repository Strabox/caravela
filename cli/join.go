package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"net"
)

func join(c *cli.Context) {

	if c.NArg() < 2 {
		log.Fatalf("Please provide the host IP address and the join node IP address")
	}

	hostIP := c.Args().Get(0)
	if net.ParseIP(hostIP) == nil {
		log.Fatalf("Please provide a valid host IP address")
	}

	joinIP := c.Args().Get(1)
	if net.ParseIP(joinIP) == nil {
		log.Fatalf("Please provide a valid join IP address")
	}

	initNode(hostIP, true, joinIP)
}
