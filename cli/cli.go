package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"path"
)

const Usage = "Caravela a fully decentralized docker cluster platform"
const Version = "0.0.1"

const Author = "André Pires"
const Email = "pardal.pires@tecnico.ulisboa.pt"

func Start() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = Usage
	app.Version = Version

	app.Author = Author
	app.Email = Email

	// Application global flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "debug, d",
			Value: "fatal",
			Usage: "Print debug traces depending on the granularity level",
		},
	}

	// Before running the user's command
	app.Before = func(context *cli.Context) error {

		switch context.String("debug") {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warning":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		case "fatal":
			log.SetLevel(log.FatalLevel)
		case "panic":
			log.SetLevel(log.PanicLevel)
		}

		logOutputFormatter := &log.TextFormatter{}
		logOutputFormatter.DisableColors = true
		log.SetFormatter(logOutputFormatter)
		log.SetOutput(os.Stdout)

		log.Infof("##################################################################")
		log.Infof("#          CARAVELA: A Cloud @ Edge                 000000       #")
		log.Infof("#            Author: %s                 00000000000     #", Author)
		log.Infof("#  Email: %s           | ||| |      #", Email)
		log.Infof("#              IST/INESC-ID                        || ||| ||     #")
		log.Infof("##################################################################")
		return nil
	}

	app.Commands = commands

	// Run the user's command
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
