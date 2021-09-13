package commands

import (
	"fmt"
	"strings"

	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/kataras/golog"
	"github.com/rs/xid"
	"github.com/urfave/cli/v2"
)

func EnqueueCMD(q queue.Driver) *cli.Command {
	return &cli.Command{
		Name:            "enqueue:cmd",
		Description:     "submit a raw shell command to the queue",
		SkipFlagParsing: true,
		Action: func(ctx *cli.Context) error {
			err := q.Enqueue(&queue.Job{
				ID:  xid.New().String(),
				Cmd: ctx.Args().Slice(),
			})

			return err
		},
	}
}

func EnqueueMacro(cfg *config.Config, q queue.Driver) *cli.Command {
	return &cli.Command{
		Name:        "enqueue:macro",
		Description: "submit a raw shell command to the queue",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "macro",
				Aliases:  []string{"m"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:    "args",
				Aliases: []string{"a"},
			},
		},

		Action: func(ctx *cli.Context) error {
			argsSlice := ctx.StringSlice("args")
			argsMap := map[string]string{}

			macro, found := cfg.LookupMacro(ctx.String("macro"))
			if !found {
				return fmt.Errorf("macro %s not found", ctx.Args().First())
			}

			for _, arg := range argsSlice {
				parts := strings.SplitN(arg, "=", 2)

				if len(parts) < 2 {
					return fmt.Errorf("the argument (%s) you supplied is invalid, it should be in the form of SOME_KEY=SOME_VALUE", arg)
				}

				argsMap[parts[0]] = parts[1]
			}

			cmd, err := macro.ParseAndSplit(argsMap)
			if err != nil {
				return err
			}

			golog.Info(
				fmt.Sprintf(
					"macro (%s) is expanded to be (%s)\n",
					macro.Name,
					strings.Join(cmd, " "),
				),
			)

			err = q.Enqueue(&queue.Job{
				Cmd: cmd,
			})

			return err
		},
	}
}
