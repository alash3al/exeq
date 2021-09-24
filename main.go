package main

import (
	"log"
	"os"

	"github.com/alash3al/exeq/internals/commands"
	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/alash3al/exeq/pkg/utils"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/golog"
	"github.com/urfave/cli/v2"

	_ "github.com/alash3al/exeq/internals/queue/drivers/rmq"
)

var (
	cfg       *config.Config
	queueConn queue.Driver
)

const (
	configFileEnvName = "EXEQ_CONFIG"
)

func init() {
	var err error

	cfg, err = config.BootFromFile(utils.Getenv(configFileEnvName, "./exeq.hcl"))
	if err != nil {
		golog.Fatal(err)
	}

	queueConn, err = queue.Open(cfg.Queue.Driver, cfg.Queue)
	if err != nil {
		golog.Fatal(err)
	}

	if cfg.Logging.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			// Either set your DSN here or set the SENTRY_DSN environment variable.
			Dsn:     cfg.Logging.SentryDSN,
			Release: "exeq",
		}); err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}

	go (func() {
		for err := range queueConn.Err() {
			golog.Error(err)

			if cfg.Logging.SentryDSN != "" {
				sentry.CaptureException(err)
			}
		}
	})()
}

func main() {
	golog.SetLevel(cfg.Logging.LogLevel)

	app := &cli.App{
		Name:                   "exeq",
		Version:                "1.0.0",
		Description:            "exeq enables you to execute shell commands in a managed scalable queue",
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,

		Commands: []*cli.Command{
			commands.QueueWork(cfg, queueConn),
			commands.EnqueueMacro(cfg, queueConn),
			commands.EnqueueCMD(queueConn),
			commands.QueueList(queueConn),
			commands.QueueStats(queueConn),
			commands.HTTPServer(cfg, queueConn),
		},
	}

	if err := app.Run(os.Args); err != nil {
		golog.Fatal(err)
	}
}
