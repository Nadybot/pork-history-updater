package main

import (
	"log/slog"
	"os"

	"pork-history-updater/internal/adapters/external/pork"
	"pork-history-updater/internal/adapters/persistence/dryrun"
	"pork-history-updater/internal/adapters/persistence/mysql"
	"pork-history-updater/internal/application"
	"pork-history-updater/pkg/config"
	"pork-history-updater/platform/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Cannot load configuration", "error", err)
		os.Exit(1)
	}

	database, err := db.New(cfg.DB)
	if err != nil {
		slog.Error("Cannot connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	var playerRepo application.PlayerRepository = mysql.NewPlayerRepository(database)
	if cfg.DryRun {
		playerRepo = dryrun.NewPlayerRepository(playerRepo)
	}

	porkClient := pork.NewClient()
	charInfoFetcher := pork.NewAdapter(porkClient)

	updater := application.NewUpdater(charInfoFetcher, playerRepo, cfg.MaxWorkers)
	updater.Run()
}
