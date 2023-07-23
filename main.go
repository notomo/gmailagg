package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/notomo/gmailagg/app"
	"github.com/urfave/cli/v2"
)

const (
	paramConfigFilePath = "config"
	paramDryRun         = "dry-run"
	paramLogDir         = "log-dir"
)

func main() {
	app := &cli.App{
		Name: "gmailagg",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     paramConfigFilePath,
				Required: true,
				Usage:    "config",
			},
			&cli.StringFlag{
				Name:     paramLogDir,
				Required: false,
				Usage:    "log directory (output log if not empty)",
			},
		},

		Commands: cli.Commands{

			{
				Name: "auth",
				Action: func(c *cli.Context) error {
					config, err := app.ReadConfig(c.String(paramConfigFilePath))
					if err != nil {
						return fmt.Errorf("read config: %w", err)
					}

					if err := app.Authorize(
						c.Context,
						config.GmailCredentialsPath,
						os.Stdout,
						os.Stdin,
						app.TokenFilePath(),
						app.LogTransport(c.String(paramLogDir), http.DefaultTransport),
					); err != nil {
						return fmt.Errorf("authorize: %w", err)
					}
					return nil
				},
			},

			{
				Name: "run",
				Action: func(c *cli.Context) error {
					config, err := app.ReadConfig(c.String(paramConfigFilePath))
					if err != nil {
						return fmt.Errorf("read config: %w", err)
					}

					var dryRunWriter io.Writer
					if c.Bool(paramDryRun) {
						dryRunWriter = os.Stdout
					}

					if err := app.Run(
						c.Context,
						config.GmailCredentialsPath,
						app.TokenFilePath(),
						config.Measurements,
						config.Influxdb.ServerURL,
						os.Getenv("INFLUXDB_TOKEN"),
						config.Influxdb.Org,
						config.Influxdb.Bucket,
						app.LogTransport(c.String(paramLogDir), http.DefaultTransport),
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
		log.Fatal(err)
	}
}
