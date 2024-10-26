package state

import (
	"context"
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"

	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
	"github.com/vs49688/accords-mirrorrer/cmd/accords-mirrorrer/config"
)

type configuration struct {
	Configuration *config.Configuration
	StateFile     string
}

func RegisterCommand(app *cli.App, globalCfg *config.Configuration) *cli.App {
	cfg := configuration{
		Configuration: globalCfg,
		StateFile:     "state.json",
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:  "state",
		Usage: "state manipulation operations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "state-file",
				Usage:       "state file",
				Value:       cfg.StateFile,
				Destination: &cfg.StateFile,
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:  "export",
				Usage: "export the state, unsetting download completion flags",
				Action: func(context *cli.Context) error {
					return exportState(context.Context, &cfg)
				},
			},
		},
	})

	return app
}

func exportState(_ context.Context, cfg *configuration) error {
	state, err := accords_mirrorrer.LoadState(cfg.StateFile)
	if err != nil {
		return err
	}

	for _, di := range state.Downloads {
		di.Completed = false
	}

	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(b)
	return err
}
