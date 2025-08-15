package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/cmd/accords-mirrorrer/archive"
	"git.vs49688.net/zane/accords-mirrorrer/cmd/accords-mirrorrer/config"
	"git.vs49688.net/zane/accords-mirrorrer/cmd/accords-mirrorrer/state"
)

func main() {
	cfg := config.Configuration{}

	app := cli.App{
		Name:    "accords-mirrorrer",
		Usage:   "Accord's Library Mirrorrer",
		Version: accords_mirrorrer.Version,
		Before: func(context *cli.Context) error {
			cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: cfg.LogLevel,
			}))

			return nil
		},
		Flags: cfg.Flags(),
	}

	archive.RegisterCommand(&app, &cfg)
	state.RegisterCommand(&app, &cfg)

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
