package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Database encapsulates the postgres connection and implements connection pooling
type Database struct {
	conn *sql.DB
	mu   sync.RWMutex
}

// New creates a new database connection with connection pooling
func New(dsn string) (*Database, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connection established")

	return &Database{conn: conn}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// GetConn returns the underlying database connection (read-only safe)
func (d *Database) GetConn() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn
}

// Exec wraps database exec with panic recovery
func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn.Exec(query, args...)
}

// Query wraps database query with panic recovery
func (d *Database) Query(query string, args ...interface{}) (*sql.Rows, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn.Query(query, args...)
}

// QueryRow wraps database query row
func (d *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn.QueryRow(query, args...)
}

// BeginTx starts a transaction
func (d *Database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.conn.BeginTx(ctx, opts)
}
