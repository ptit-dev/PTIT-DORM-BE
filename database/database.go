package database

import (
	"database/sql"
	"Backend_Dorm_PTIT/config"
	"Backend_Dorm_PTIT/logger"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDatabase(cfg *config.DatabaseConfig) error {
	var err error

	logger.Info().
		Str("host", cfg.Host).
		Str("port", cfg.Port).
		Str("database", cfg.Name).
		Str("schema", cfg.Schema).
		Msg("Initializing database connection")

	dsn := cfg.GetDSN()

	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to open database connection")
		return fmt.Errorf("failed to open database: %w", err)
	}

	schema := cfg.Schema
	_, err = DB.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schema))
	if err != nil {
		logger.Error().Err(err).Str("schema", schema).Msg("Failed to create schema")
		return fmt.Errorf("failed to create schema: %w", err)
	}

	_, err = DB.Exec(fmt.Sprintf(`SET search_path TO %s`, schema))
	if err != nil {
		logger.Error().Err(err).Str("schema", schema).Msg("Failed to set schema")
		return fmt.Errorf("failed to set schema: %w", err)
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		logger.Error().Err(err).Msg("Failed to ping database")
		return fmt.Errorf("failed to connect to database: %w", err)
	}


	logger.Info().
		Msg("Database connected successfully")



	return nil
}

func GetDB() *sql.DB {
	return DB
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
