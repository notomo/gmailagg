package main

import (
	"fmt"
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
)

func main() {
	app.SetupLogger()

	tokenFilePath, err := app.GetDefaultTokenPath()
	if err != nil {
		slog.Default().Error(err.Error())
		os.Exit(1)
	}

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
				Value:    tokenFilePath,
			},
		},

		Commands: cli.Commands{

			{
				Name: "auth",
				Action: func(c *cli.Context) error {
					if err := app.Authorize(
						c.Context,
						os.Getenv(GMAILAGG_GMAIL_CREDENTIALS),
						c.String(paramTokenPath),
						browser.New(os.Stdout, os.Stderr),
						c.Duration(paramTimeout),
						c.Uint(paramPort),
						app.LogTransport(c.String(paramLogDir), http.DefaultTransport),
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
					config, err := app.ReadConfig(c.String(paramConfigFilePath))
					if err != nil {
						return fmt.Errorf("read config: %w", err)
					}

					if err := app.Run(
						c.Context,
						os.Getenv(GMAILAGG_GMAIL_CREDENTIALS),
						c.String(paramTokenPath),
						config.Measurements,
						app.LogTransport(c.String(paramLogDir), http.DefaultTransport),
						os.Stdout,
					); err != nil {
						return fmt.Errorf("run: %w", err)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Default().Error(err.Error())
		os.Exit(1)
	}
}
