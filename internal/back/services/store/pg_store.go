package store

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/golang-migrate/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore initializes the connection pool and returns a new PostgresStore instance.
func NewPostgresStore(ctx context.Context, connString string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	store := &PostgresStore{pool: pool}
	err = store.runMigrations(connString)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *PostgresStore) runMigrations(connString string) error {
	mPath := getMigrationPath()
	m, err := migrate.New(mPath, connString)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func getMigrationPath() string {
	_, currentFilePath, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFilePath)
	migrationsPath := filepath.Join(currentDir, "migrations")
	migrationsPath = filepath.ToSlash(migrationsPath)
	return "file://" + migrationsPath
}

// CreateTask adds a task to the queue, userfiles, outbox.
func (s *PostgresStore) CreateTask(task model.StoreTask) error {
	return nil
}

// GetState retrieves the user state based on the userId.
func (s *PostgresStore) GetState(userId int64) ([]model.UserItem, error) {
	return nil, nil
}

// Close closes the connection pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}
