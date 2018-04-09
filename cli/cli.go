package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/version"
	"github.com/urfave/cli"
	"os"
	"path"
)

func Run() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = AppUsage
	app.Version = version.Version
	app.Author = Author
	app.Email = Email

	// Application global flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "debug, d",
			Value: "fatal",
			Usage: "Log traces depending on the granularity level",
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
		// Set the format of the log text and the place to write
		logOutputFormatter := &log.TextFormatter{}
		logOutputFormatter.DisableColors = true
		logOutputFormatter.DisableTimestamp = true
		log.SetFormatter(logOutputFormatter)
		log.SetOutput(os.Stdout)
		return nil
	}

	app.Commands = commands

	// Run the user's command
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
