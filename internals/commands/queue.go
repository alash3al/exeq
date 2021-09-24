package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

func QueueWork(cfg *config.Config, q queue.Driver) *cli.Command {
	return &cli.Command{
		Name:  "queue:work",
		Usage: "start the queue worker(s)",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "workers",
				Aliases:     []string{"w"},
				Usage:       "the workers count, this will override the configuration file value",
				DefaultText: strconv.Itoa(cfg.Queue.WorkersCount),
				EnvVars:     []string{"EXEQ_WORKERS_COUNT"},
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.IsSet("workers") {
				cfg.Queue.WorkersCount = ctx.Int("workers")
			}

			if err := cfg.SetupMounts(); err != nil {
				return err
			}

			return q.ListenAndConsume()
		},
	}
}

func QueueList(q queue.Driver) *cli.Command {
	return &cli.Command{
		Name:  "queue:jobs",
		Usage: "list the jobs",
		Action: func(ctx *cli.Context) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Created", "Status", "Command"})
			table.SetAutoWrapText(false)
			table.SetRowSeparator("")
			table.SetColumnSeparator("")
			table.SetCenterSeparator("")
			table.SetTablePadding("\t")
			table.SetHeaderLine(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			data, err := q.List()
			if err != nil {
				return err
			}

			for _, v := range data {
				status := "pending"

				if v.IsRunning() {
					status = fmt.Sprintf("up for %s", time.Since(v.StartedAt).Round(time.Minute))
				} else if v.HasFinished() {
					exitMsg := "Ok"
					if v.Error != "" {
						exitMsg = v.Error
					}
					status = fmt.Sprintf(
						"exited (%s) %s ago",
						exitMsg,
						time.Since(v.FinishedAt).Round(time.Minute),
					)
				}

				table.Append([]string{
					v.ID,
					time.Since(v.EnqueuedAt).Round(time.Minute).String(),
					status,
					strings.Join(v.Cmd, " "),
				})
			}

			table.Render()

			return nil
		},
	}
}

func QueueStats(q queue.Driver) *cli.Command {
	return &cli.Command{
		Name:  "queue:stats",
		Usage: "show queue stats",
		Action: func(ctx *cli.Context) error {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"PENDING", "RUNNING", "FAILED", "RETRIES", "SUCCEEDED"})
			table.SetAutoWrapText(false)
			table.SetRowSeparator("")
			table.SetColumnSeparator("")
			table.SetCenterSeparator("")
			table.SetTablePadding("\t")
			table.SetHeaderLine(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			data, err := q.Stats()
			if err != nil {
				return err
			}

			table.Append([]string{
				fmt.Sprintf("%d", data.Pending),
				fmt.Sprintf("%d", data.Running),
				fmt.Sprintf("%d", data.Failed),
				fmt.Sprintf("%d", data.Retries),
				fmt.Sprintf("%d", data.Succeeded),
			})

			table.Render()

			return nil
		},
	}
}
