package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
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
func (s *PostgresStore) CreateTasks(ctx context.Context, tasks []model.StoreTask) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	insertQueueStmt := `INSERT INTO queue (order_num) VALUES (DEFAULT) RETURNING id`
	insertUserFileStmt := `INSERT INTO userfiles (queue_id, user_id, file_name, src_file_url, src_file_key, dest_file_url, dest_file_key, state) 
						   VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	insertOutboxStmt := `INSERT INTO outbox (payload, status, idempotency_key) 
						 VALUES ($1, 'PENDING', $2)`

	for _, task := range tasks {
		var queueID int
		err = tx.QueryRow(ctx, insertQueueStmt).Scan(&queueID)
		if err != nil {
			return fmt.Errorf("tx.QueryRow insertQueueStmt %w", err)
		}

		var fileID int64
		err = tx.QueryRow(ctx,
			insertUserFileStmt,
			queueID, task.UserID, task.FileName, task.SrcFileURL, task.SrcFileKey, task.DestFileURL,
			task.DestFileKey, "PENDING").Scan(&fileID)
		if err != nil {
			return fmt.Errorf("tx.QueryRow insertUserFileStmt %w", err)
		}

		payload := cmodel.BrokerMessage{
			SrcFileURL:    task.SrcFileURL,
			DestFileURL:   task.DestFileURL,
			TaskId:        fileID,
			FileExtension: filepath.Ext(task.FileName),
		}
		plJson, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("json.Marshal %w", err)
		}

		_, err = tx.Exec(ctx, insertOutboxStmt, plJson, strconv.FormatInt(fileID, 10))
		if err != nil {
			return fmt.Errorf("tx.Exec insertOutboxStmt %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit %w", err)
	}

	return nil
}

// GetState retrieves the user state based on the userId.
func (s *PostgresStore) GetState(ctx context.Context, userId int64) ([]model.UserItem, error) {
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

	rows, err := s.pool.Query(ctx, query, userId)
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
func (s *PostgresStore) CreateUser(ctx context.Context) (int64, error) {
	query := `
        INSERT INTO users (created_at)
        VALUES (NOW())
        RETURNING id;
    `
	var userID int64

	err := s.pool.QueryRow(ctx, query).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *PostgresStore) SendTasksToBroker(ctx context.Context, sendFunc func(items []model.OutboxItem) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pool.Begin %w", err)
	}

	rows, err := tx.Query(ctx, `
        SELECT id, payload, idempotency_key
        FROM outbox
        WHERE status = 'PENDING'
        ORDER BY created_at
        LIMIT 100
        FOR UPDATE SKIP LOCKED`)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("outbox select query %w", err)
	}
	defer rows.Close()

	var items []model.OutboxItem
	var ids []int

	for rows.Next() {
		var id int
		var payload string
		var idempotencyKey string

		if err := rows.Scan(&id, &payload, &idempotencyKey); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("outbox select query rows.Scan %w", err)
		}

		item := model.OutboxItem{Payload: payload, IdKey: idempotencyKey}
		items = append(items, item)
		ids = append(ids, id)
	}

	if len(items) > 0 {
		sendFunc(items)
		updateQuery := `
		UPDATE outbox
		SET status = 'SENT', processed_at = NOW()
		WHERE id = ANY($1)`
		_, err = tx.Exec(ctx, updateQuery, pq.Array(ids))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("outbox update query %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("tx.Commit %w", err)
		} else {
			return err
		}
	} else {
		tx.Rollback(ctx)
	}

	return nil
}

// FinishTasks обновляет состояние userfiles и удаляет соответствующую строку из queue.
func (s *PostgresStore) FinishTasks(ctx context.Context, msgs []cmodel.BrokerMessageResult) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	for _, msg := range msgs {
		var queueID int

		getQueueIDStmt := `
            SELECT queue_id
            FROM userfiles
            WHERE id = $1
        `
		err = tx.QueryRow(ctx, getQueueIDStmt, msg.TaskId).Scan(&queueID)
		if err != nil {
			return fmt.Errorf("ошибка при получении queue_id: %w", err)
		}

		updateUserFilesStmt := `
            UPDATE userfiles
            SET state = $1, queue_id = NULL
            WHERE id = $2
        `
		_, err = tx.Exec(ctx, updateUserFilesStmt, msg.Result, msg.TaskId)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении userfiles: %w", err)
		}

		// Удаляем строку из таблицы queue, используя queue_id
		deleteQueueStmt := `
            DELETE FROM queue
            WHERE id = $1
        `
		_, err = tx.Exec(ctx, deleteQueueStmt, queueID)
		if err != nil {
			return fmt.Errorf("ошибка при удалении из queue: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit %w", err)
	}

	return nil
}

// Close closes the connection pool.
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}
