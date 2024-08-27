package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/notomo/gmailagg/app"
	"github.com/notomo/gmailagg/pkg/browser"
	"github.com/urfave/cli/v2"
)

const (
	paramConfigFilePath = "config"
	paramDryRun         = "dry-run"
	paramLogDir         = "log-dir"
	paramTokenPath      = "token"
	paramTimeout        = "timeout"
	paramPort           = "port"
)

const (
	GMAILAGG_GMAIL_CREDENTIALS = "GMAILAGG_GMAIL_CREDENTIALS"
	INFLUXDB_TOKEN             = "INFLUXDB_TOKEN"
)

func main() {
	app.SetupLogger()

	app := &cli.App{
		Name: "gmailagg",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     paramConfigFilePath,
				Required: false,
				Usage:    "config",
			},
			&cli.StringFlag{
				Name:     paramLogDir,
				Required: false,
				Usage:    "log directory (output log if not empty)",
			},
			&cli.StringFlag{
				Name:     paramTokenPath,
				Required: false,
				Usage:    "token file path",
				Value:    "gs://gmailagg-oauth/token.json",
			},
		},

		Commands: cli.Commands{

			{
				Name: "auth",
				Action: func(c *cli.Context) error {
					baseTransport := app.LogTransport(c.String(paramLogDir), http.DefaultTransport)

					opener := browser.New(os.Stdout, os.Stderr)

					if err := app.Authorize(
						c.Context,
						os.Getenv(GMAILAGG_GMAIL_CREDENTIALS),
						c.String(paramTokenPath),
						opener,
						c.Duration(paramTimeout),
						c.Uint(paramPort),
						baseTransport,
						c.Bool(paramDryRun),
					); err != nil {
						return fmt.Errorf("authorize: %w", err)
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     paramDryRun,
						Required: false,
						Value:    false,
						Usage:    "dry run",
					},
					&cli.DurationFlag{
						Name:     paramTimeout,
						Required: false,
						Value:    3 * time.Minute,
						Usage:    "user operation's timeout",
					},
					&cli.UintFlag{
						Name:     paramPort,
						Required: false,
						Value:    0,
						Usage:    "redirect uri port",
					},
				},
			},

			{
				Name: "run",
				Action: func(c *cli.Context) error {
					ctx := c.Context
					baseTransport := app.LogTransport(c.String(paramLogDir), http.DefaultTransport)
					config, err := app.ReadConfig(
						ctx,
						c.String(paramConfigFilePath),
						baseTransport,
					)
					if err != nil {
						return fmt.Errorf("read config: %w", err)
					}

					var dryRunWriter io.Writer
					if c.Bool(paramDryRun) {
						dryRunWriter = os.Stdout
					}

					if err := app.Run(
						ctx,
						os.Getenv(GMAILAGG_GMAIL_CREDENTIALS),
						c.String(paramTokenPath),
						config.Measurements,
						config.Influxdb.ServerURL,
						os.Getenv(INFLUXDB_TOKEN),
						config.Influxdb.Org,
						config.Influxdb.Bucket,
						baseTransport,
						dryRunWriter,
					); err != nil {
						return fmt.Errorf("run: %w", err)
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     paramDryRun,
						Required: false,
						Value:    false,
						Usage:    "dry run",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Default().Error(err.Error())
		os.Exit(1)
	}
}
