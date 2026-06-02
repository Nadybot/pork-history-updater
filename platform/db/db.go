package db

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Config holds all connection parameters.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Type     string
}

// New creates a new database connection.
func New(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4\u0026parseTime=True\u0026loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	db, err := sqlx.Connect(cfg.Type, dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Configure connection
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}
