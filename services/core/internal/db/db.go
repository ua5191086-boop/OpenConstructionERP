package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB wraps *sqlx.DB with helper methods
type DB struct {
	*sqlx.DB
}

// StdDB returns the underlying *sql.DB for legacy handlers
func (d *DB) StdDB() *sql.DB {
	return d.DB.DB
}

// New creates a new database connection
func New() (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "oce"),
		getEnv("DB_PASSWORD", "oce_secret"),
		getEnv("DB_NAME", "oce_erp"),
	)

	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("[DB] connected to PostgreSQL (sqlx)")
	return &DB{conn}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
