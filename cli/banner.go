package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

/*
Prints the banner of the CARAVELA system into the log.
*/
func printBanner(_ *cli.Context) error {
	log.Infof("##################################################################")
	log.Infof("#          CARAVELA: A Cloud @ Edge                 000000       #")
	log.Infof("#            Author: %s                 00000000000     #", Author)
	log.Infof("#  Email: %s           | ||| |      #", Email)
	log.Infof("#              IST/INESC-ID                        || ||| ||     #")
	log.Infof("##################################################################")
	return nil
}
