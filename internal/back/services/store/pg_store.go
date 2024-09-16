package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
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
	migrationsPath := filepath.Join(currentDir, "migration")
	migrationsPath = filepath.ToSlash(migrationsPath)
	return "file://" + migrationsPath
}

// CreateTasks adds a tasks to the queue, userfiles, outbox.
func (s *PostgresStore) CreateTasks(tasks []model.StoreTask) error {
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(context.Background())
		} else {
			err = tx.Commit(context.Background())
		}
	}()

	insertQueueStmt := `INSERT INTO queue (order_num) VALUES (DEFAULT) RETURNING id`
	insertUserFileStmt := `INSERT INTO userfiles (queue_id, user_id, file_name, src_file_url, src_file_key, dest_file_url, dest_file_key, state) 
						   VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	insertOutboxStmt := `INSERT INTO outbox (payload, status) 
						 VALUES ($1, 'PENDING')`

	for _, task := range tasks {
		var queueID int
		err = tx.QueryRow(context.Background(), insertQueueStmt).Scan(&queueID)
		if err != nil {
			return err
		}

		var taskId int
		err = tx.QueryRow(context.Background(),
			insertUserFileStmt,
			queueID, task.UserID, task.FileName, task.SrcFileURL, task.SrcFileKey, task.DestFileURL, task.DestFileKey, "PENDING").Scan(&taskId)
		if err != nil {
			return err
		}

		payload := model.BrokerMessage{
			SrcFileURL:  task.SrcFileURL,
			DestFileURL: task.DestFileURL,
			TaskId:      int64(taskId),
		}
		plJson, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		_, err = tx.Exec(context.Background(),
			insertOutboxStmt,
			plJson)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetState retrieves the user state based on the userId.
func (s *PostgresStore) GetState(userId int64) ([]model.UserItem, error) {
	query := `
	WITH queue_positions AS (
		SELECT 
			id,
			ROW_NUMBER() OVER (ORDER BY id) AS position 
		FROM queue
	)
	SELECT
		uf.order_num,                 
		uf.file_name,             
		uf.dest_file_url,             
		COALESCE(qp.position, -1),    
		uf.state                      
	FROM
		userfiles uf
	LEFT JOIN
		queue_positions qp ON uf.queue_id = qp.id
	WHERE
		uf.user_id = $1
	ORDER BY
		uf.order_num
	`

	rows, err := s.pool.Query(context.Background(), query, userId)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса состояния пользователя: %w", err)
	}
	defer rows.Close()

	var userItems []model.UserItem

	for rows.Next() {
		var item model.UserItem
		err := rows.Scan(&item.Order, &item.FileName, &item.Link, &item.QueuePosition, &item.Status)
		if item.Status == "PENDING" { // TODO to const
			item.Link = ""
		}
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки: %w", err)
		}
		userItems = append(userItems, item)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %w", rows.Err())
	}

	return userItems, nil
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
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}
