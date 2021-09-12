package commands

import (
	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/urfave/cli/v2"
)

func QueueWork(cfg *config.Config, q *queue.Queue) *cli.Command {
	return &cli.Command{
		Name:        "queue:work",
		Description: "start the queue worker(s)",
		Action: func(ctx *cli.Context) error {
			if err := cfg.SetupMounts(); err != nil {
				return err
			}

			return q.ListenAndConsume()
		},
	}
}
