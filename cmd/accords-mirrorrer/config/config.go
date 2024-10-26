package config

import (
	"log/slog"

	"github.com/urfave/cli/v2"
	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
)

type Configuration struct {
	LogLevel  slog.Level
	UserAgent string

	Logger *slog.Logger
}

func DefaultConfiguration() Configuration {
	return Configuration{
		LogLevel:  slog.LevelInfo,
		UserAgent: accords_mirrorrer.UserAgent,
		Logger:    slog.Default(),
	}
}

func (cfg *Configuration) Flags() []cli.Flag {
	def := DefaultConfiguration()

	return []cli.Flag{
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "log level",
			Value: def.LogLevel.String(),
			Action: func(context *cli.Context, s string) error {
				return cfg.LogLevel.UnmarshalText([]byte(s))
			},
		},
		&cli.StringFlag{
			Name:        "user-agent",
			Usage:       "override http user agent",
			Value:       def.UserAgent,
			Destination: &cfg.UserAgent,
		},
	}
}
