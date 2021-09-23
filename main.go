package main

import (
	"os"

	"github.com/alash3al/exeq/internals/commands"
	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/alash3al/exeq/pkg/utils"
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

	go (func() {
		for err := range queueConn.Err() {
			golog.Error(err)
		}
	})()
}

func main() {
	golog.SetLevel(cfg.LogLevel)

	app := &cli.App{
		Name:        "exeq",
		Version:     "1.0.0",
		Description: "exeq enables you to execute shell commands in a managed scalable queue",

		Commands: []*cli.Command{
			commands.QueueWork(cfg, queueConn),
			commands.EnqueueMacro(cfg, queueConn),
			commands.EnqueueCMD(queueConn),
		},
	}

	if err := app.Run(os.Args); err != nil {
		golog.Fatal(err)
	}
}
