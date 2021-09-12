package main

import (
	"log"
	"os"

	"github.com/alash3al/exeq/internals/commands"
	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/alash3al/exeq/pkg/utils"
	"github.com/urfave/cli/v2"
)

var (
	cfg       config.Config
	queueMngr *queue.Queue
)

const (
	configFileEnvName = "EXEQ_CONFIG"
)

func init() {
	var err error

	cfg, err = config.BootFromFile(utils.Getenv(configFileEnvName, "./exeq.hcl"))
	if err != nil {
		log.Fatal(err)
	}

	queueMngr, err = queue.New(&cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := &cli.App{
		Name:        "exeq",
		Version:     "1.0.0",
		Description: "exeq enables you to execute shell commands in a managed scalable queue",

		Commands: []*cli.Command{
			commands.QueueWork(&cfg, queueMngr),
			commands.EnqueueMacro(&cfg, queueMngr),
			commands.EnqueueCMD(queueMngr),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
