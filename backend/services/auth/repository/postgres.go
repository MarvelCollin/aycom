package repository

import (
	"database/sql"
	"fmt"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgresConnection creates a new connection to the PostgreSQL database
func NewPostgresConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create the auth_tokens table if it doesn't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// createTables creates the necessary tables if they don't exist
func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS auth_tokens (
		id SERIAL PRIMARY KEY,
		user_id TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		revoked BOOLEAN NOT NULL DEFAULT FALSE
	);
	
	CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_id ON auth_tokens(user_id);
	CREATE INDEX IF NOT EXISTS idx_auth_tokens_refresh_token ON auth_tokens(refresh_token);
	`

	_, err := db.Exec(query)
	return err
}
