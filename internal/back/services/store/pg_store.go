package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // needs for init
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewPostgresStore initializes the connection pool and returns a new PostgresStore instance.
func NewPostgresStore(ctx context.Context, connString string, log *slog.Logger) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	store := &PostgresStore{
		pool: pool,
		log:  log,
	}
	err = store.runMigrations(connString, log)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *PostgresStore) runMigrations(connString string, log *slog.Logger) error {
	mPath := getMigrationPath()
	log.Info("runMigrations", "mPath", mPath)

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
	migrationsPath := filepath.Join(currentDir, "migration")
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

// GetState retrieves the user state based on the userId.
func (s *PostgresStore) CreateUser() (int64, error) {
	query := `
        INSERT INTO users (created_at)
        VALUES (NOW())
        RETURNING id;
    `
	var userID int64

	err := s.pool.QueryRow(context.Background(), query).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// Close closes the connection pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}
