package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alash3al/exeq/internals/config"
	"github.com/alash3al/exeq/internals/queue"
	"github.com/rs/xid"
	"github.com/urfave/cli/v2"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	sentryecho "github.com/getsentry/sentry-go/echo"
)

func HTTPServer(cfg *config.Config, queueConn queue.Driver) *cli.Command {
	return &cli.Command{
		Name:  "serve:http",
		Usage: "start the http listener",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Aliases:     []string{"a"},
				Usage:       "the http listen address in the following formats hostname:port, :port",
				DefaultText: cfg.HTTPServer.ListenAddr,
				EnvVars:     []string{"EXEQ_HTTP_LISTEN_ADDR"},
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.String("addr") != "" {
				cfg.HTTPServer.ListenAddr = ctx.String("addr")
			}

			return listenAndServe(cfg, queueConn)
		},
	}
}

func listenAndServe(cfg *config.Config, queueConn queue.Driver) error {
	e := echo.New()
	e.HideBanner = true
	e.Debug = cfg.Logging.LogLevel == "debug"

	if cfg.HTTPServer.EnableAccessLogs {
		e.Use(middleware.Logger())
	}

	e.Use(middleware.Recover())

	if cfg.Logging.SentryDSN != "" {
		e.Use(sentryecho.New(sentryecho.Options{}))
	}

	e.Use(middleware.RemoveTrailingSlash())

	e.GET("/", func(c echo.Context) error {
		stats, _ := queueConn.Stats()

		return c.JSON(200, echo.Map{
			"success": true,
			"message": "welcome to exeq web interface",
			"payload": echo.Map{
				"stats": stats,
			},
		})
	})

	e.GET("/metrics", func(c echo.Context) error {
		stats, err := queueConn.Stats()
		if err != nil {
			return c.JSON(500, echo.Map{
				"success": true,
				"error":   err.Error(),
			})
		}

		metrics := []string{}

		metrics = append(metrics, []string{
			`# TYPE exeq_jobs_pending gauge`,
			`# HELP exeq_jobs_pending Number of pending jobs.`,
			fmt.Sprintf(`exeq_jobs_pending %d`, stats.Pending),
		}...)

		metrics = append(metrics, []string{
			`# TYPE exeq_jobs_failed counter`,
			`# HELP exeq_jobs_failed Number of failed jobs.`,
			fmt.Sprintf(`exeq_jobs_failed %d`, stats.Failed),
		}...)

		metrics = append(metrics, []string{
			`# TYPE exeq_jobs_retries counter`,
			`# HELP exeq_jobs_retries Number of retries.`,
			fmt.Sprintf(`exeq_jobs_failed %d`, stats.Retries),
		}...)

		metrics = append(metrics, []string{
			`# TYPE exeq_jobs_running gauge`,
			`# HELP exeq_jobs_running Number of currently running jobs.`,
			fmt.Sprintf(`exeq_jobs_failed %d`, stats.Running),
		}...)

		metrics = append(metrics, []string{
			`# TYPE exeq_jobs_succeeded counter`,
			`# HELP exeq_jobs_succeeded Number of succeeded jobs.`,
			fmt.Sprintf(`exeq_jobs_failed %d`, stats.Succeeded),
		}...)

		return c.String(200, strings.Join(metrics, "\n"))
	})

	e.POST("/enqueue/:macro", func(c echo.Context) error {
		macro, exists := cfg.LookupMacro(c.Param("macro"))
		if !exists {
			return c.JSON(404, echo.Map{
				"success": false,
				"error":   "macro not found",
			})
		}

		maxExecTime, _ := time.ParseDuration(macro.MaxExecTime)

		var argsMap map[string]interface{}

		if err := json.NewDecoder(c.Request().Body).Decode(&argsMap); err != nil {
			return c.JSON(400, echo.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		argsNormalized := map[string]string{}
		for k, v := range argsMap {
			argsNormalized[k] = fmt.Sprintf("%v", v)
		}

		cmd, err := macro.ParseAndSplit(argsNormalized)
		if err != nil {
			return c.JSON(500, echo.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		job := &queue.Job{
			ID:          xid.New().String(),
			Cmd:         cmd,
			MaxExecTime: maxExecTime,
		}

		if err := queueConn.Enqueue(job); err != nil {
			return c.JSON(500, echo.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		return c.JSON(204, echo.Map{
			"success": true,
			"message": "enqueued",
			"payload": echo.Map{
				"job_id": job.ID,
			},
		})
	})

	return e.Start(cfg.HTTPServer.ListenAddr)
}
