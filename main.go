package main

import (
	"log"
	"os"

	"github.com/alash3al/exeq/config"
	"github.com/alash3al/exeq/queue"
	"github.com/urfave/cli/v2"
)

var (
	configs   config.Config
	queueMngr *queue.Queue
)

func main() {
	app := &cli.App{
		Name:        "exeq",
		Version:     "1.0.0",
		Description: "exeq enables you to execute shell commands in queues",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "the configuriations filename",
				Value:   "./exeq.hcl",
				Aliases: []string{"c"},
				EnvVars: []string{"EXEQ_CONFIG_FILENAME"},
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			configs, err = config.LoadFromFile(ctx.String("config"))
			if err != nil {
				return
			}

			queueMngr, err = queue.New(&configs)
			if err != nil {
				return
			}

			return queueMngr.ListenAndConsume()
		},
		Before: func(ctx *cli.Context) (err error) {
			configs, err = config.LoadFromFile(ctx.String("config"))
			if err != nil {
				return
			}

			queueMngr, err = queue.New(&configs)

			return
		},
		Commands: []*cli.Command{
			{
				Name: "enqueue:cmd",
				Action: func(ctx *cli.Context) error {
					_, err := queueMngr.Enqueue(&queue.Job{
						Cmd: ctx.Args().Slice(),
					})

					return err
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
