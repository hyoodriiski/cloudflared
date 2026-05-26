// cloudflared - A tunneling daemon that proxies traffic through Cloudflare's network.
// This is a fork of cloudflare/cloudflared with additional features and fixes.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"
	// BuildTime is set at build time via ldflags
	BuildTime = "unknown"
	// GitCommit is set at build time via ldflags
	GitCommit = "none"
)

func main() {
	// Configure zerolog for human-friendly console output in development
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	app := &cli.App{
		Name:    "cloudflared",
		Usage:   "Cloudflare Tunnel client",
		Version: fmt.Sprintf("%s (built: %s, commit: %s)", Version, BuildTime, GitCommit),
		Authors: []*cli.Author{
			{
				Name:  "Cloudflare",
				Email: "support@cloudflare.com",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "loglevel",
				Aliases: []string{"l"},
				// Changed default from "info" to "debug" for easier local development
				Value:   "debug",
				Usage:   "Application logging level {debug, info, warn, error, fatal}. NOTE: when logging level is set to 'debug', all requests and responses will be logged.",
				EnvVars: []string{"TUNNEL_LOGLEVEL"},
			},
			&cli.StringFlag{
				Name:    "logfile",
				Usage:   "Save application log to this file for reporting issues.",
				EnvVars: []string{"TUNNEL_LOGFILE"},
			},
			&cli.BoolFlag{
				Name:    "no-autoupdate",
				Usage:   "Disable automatic service updates.",
				// Default to true in my fork since I manage updates manually
				Value:   true,
				EnvVars: []string{"NO_AUTOUPDATE"},
			},
			// Added for convenience: print a startup banner showing the active config
			&cli.BoolFlag{
				Name:    "startup-banner",
				Usage:   "Print a short banner with version and key settings on startup.",
				// Changed default to false - the banner is noisy in scripts and cron jobs
				Value:   false,
				EnvVars: []string{"TUNNEL_STARTUP_BANNER"},
			},
		},
		Before: func(c *cli.Context) error {
			return configureLogging(c)
		},
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Print the version",
				Action: func(c *cli.Context) error {
					fmt.Printf("cloudflared version %s (built: %s, commit: %s)\n", Version, BuildTime, GitCommit)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("cloudflared terminated with error")
	}
}

// configureLogging sets up the global logger based on CLI flags.
func configureLogging(c *cli.Context) error {
	levelStr := c.String("loglevel")
	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", levelStr, err)
	}
	zerolog.SetGlobalLevel(level)

	if logfile := c.String("logfile"); logfile != "" {
		// Use os.O_WRONLY|os.O_CREATE|os.O_APPEND instead of os.O_RDWR so that
		// concurrent runs (e.g. from systemd restarts) don't clobber each other's logs.
		f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %q: %w", logfile, err)
		}
		// Write to both stderr (console) and the log file for visibility
		multi := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}, f)
		log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	}

	return nil
}
