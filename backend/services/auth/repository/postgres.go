package repository

import (
	"database/sql"
	"fmt"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// ConnectPostgres loads database config and creates a new connection
func ConnectPostgres() (*sql.DB, error) {
	// Load database configuration
	dbConfig := config.LoadDatabaseConfig()

	// Create a new database connection
	return NewPostgresConnection(dbConfig)
}

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
	if err := CreateTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// createTables creates the necessary tables if they don't exist
// Exported now to be callable from service layer during initialization/status check
func CreateTables(db *sql.DB) error {
	// Create auth_tokens table
	authTokensQuery := `
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

	// Create users table
	usersQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		gender TEXT NOT NULL,
		date_of_birth TIMESTAMP NOT NULL,
		profile_picture_path TEXT,
		banner_path TEXT,
		security_question TEXT NOT NULL,
		security_answer_hash TEXT NOT NULL,
		subscribe_newsletter BOOLEAN NOT NULL DEFAULT FALSE,
		email_verified BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`

	// Create verification_codes table
	verificationCodesQuery := `
	CREATE TABLE IF NOT EXISTS verification_codes (
		email TEXT PRIMARY KEY,
		code TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	`

	// Execute all queries in a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute auth_tokens table creation
	_, err = tx.Exec(authTokensQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create auth_tokens table: %w", err)
	}

	// Execute users table creation
	_, err = tx.Exec(usersQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Execute verification_codes table creation
	_, err = tx.Exec(verificationCodesQuery)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create verification_codes table: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
