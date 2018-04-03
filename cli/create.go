package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"net"
)

func create(c *cli.Context) {

	if c.NArg() < 1 {
		log.Fatalf("Please provide the host IP address")
	}

	hostIP := c.Args().Get(0)
	if net.ParseIP(hostIP) == nil {
		log.Fatalf("Invalid host IP address")
	}

	initNode(hostIP, false, "")
}
