package config

import (
	"fmt"
	"os"
	"pork-history-updater/platform/db"
	"strconv"

	"github.com/spf13/pflag"
)

type Config struct {
	DB         db.Config
	MaxWorkers int
	DryRun     bool
}

// Load reads configuration from command-line flags and environment variables.
func Load() (Config, error) {
	var err error
	dryRunFlag := pflag.BoolP("dry-run", "d", false, "Run without changing any data")
	maxWorkersFlag := pflag.Int("max-workers", 0, "Number of simultaneous HTTP-connections to PORK (0 = use default)")
	dbHostFlag := pflag.String("db-host", "", "Database host")
	dbPortFlag := pflag.Int("db-port", 0, "Database port")
	dbUserFlag := pflag.String("db-user", "", "Database user")
	dbPassFlag := pflag.String("db-pass", "", "Database password")
	dbNameFlag := pflag.String("db-name", "", "Database name")
	dbTypeFlag := pflag.String("db-type", "", "Database type")

	pflag.Parse()

	port := *dbPortFlag
	if port == 0 {
		port, err = strconv.Atoi(getEnv("DB_PORT", "3306"))
		if err != nil {
			return Config{}, fmt.Errorf("Invalid DB_PORT: %w", err)
		}
	}
	maxWorkers := *maxWorkersFlag
	if maxWorkers == 0 {
		maxWorkers, err = strconv.Atoi(getEnv("MAX_WORKERS", "5"))
		if err != nil {
			return Config{}, fmt.Errorf("Invalid MAX_WORKERS: %w", err)
		}
	}
	dryRun := *dryRunFlag
	if !dryRun {
		dryRun, err = strconv.ParseBool(getEnv("DRY_RUN", "0"))
		if err != nil {
			return Config{}, fmt.Errorf("Invalid DRY_RUN: %w", err)
		}
	}
	return Config{
		DB: db.Config{
			Host:     overrideEnv("DB_HOST", *dbHostFlag, "localhost"),
			Port:     port,
			User:     overrideEnv("DB_USER", *dbUserFlag, "porkhist"),
			Password: overrideEnv("DB_PASSWORD", *dbPassFlag, ""),
			Database: overrideEnv("DB_NAME", *dbNameFlag, "porkhist"),
			Type:     overrideEnv("DB_TYPE", *dbTypeFlag, "mysql"),
		},
		MaxWorkers: maxWorkers,
		DryRun:     dryRun,
	}, nil
}

// getEnv returns the value of the environment variable or the fallback.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// overrideEnv returns the flag value if set, otherwise falls back to env or default.
func overrideEnv(key string, flagValue string, fallback string) string {
	if flagValue != "" {
		return flagValue
	}
	return getEnv(key, fallback)
}
