package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// PostgresStore represents the store implementation for PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool // Connection pool for PostgreSQL.
	log  *slog.Logger  // Logger for logging operations and errors.
}

// NewPostgresStore initializes the connection pool and runs database migrations.
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

// runMigrations executes database migrations to ensure the schema is up to date.
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

// getMigrationPath constructs the file path to the migration files.
func getMigrationPath() string {
	_, currentFilePath, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFilePath)
	migrationsPath := filepath.Join(currentDir, "migration")
	migrationsPath = filepath.ToSlash(migrationsPath)
	return "file://" + migrationsPath
}

// CreateTasks inserts tasks into the queue, user files, and outbox tables.
func (s *PostgresStore) CreateTasks(ctx context.Context, tasks []model.StoreTask) ([]model.QueueItem, []UserFileItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("pool.Begin %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	insertQueueStmt := `INSERT INTO queue (order_num) VALUES (DEFAULT) RETURNING id, order_num`
	insertUserFileStmt := `INSERT INTO userfiles (queue_id, user_id, file_name, src_file_url, src_file_key, dest_file_url, dest_file_key, state) 
						   VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`
	insertOutboxStmt := `INSERT INTO outbox (payload, status, idempotency_key) 
						 VALUES ($1, 'PENDING', $2)`

	var queueItems []model.QueueItem
	var userItems []UserFileItem

	for _, task := range tasks {
		var queueID int64
		var orderNum int64
		err = tx.QueryRow(ctx, insertQueueStmt).Scan(&queueID, &orderNum)
		if err != nil {
			return nil, nil, fmt.Errorf("tx.QueryRow insertQueueStmt %w", err)
		}

		var fileID int64
		var createdAt time.Time
		err = tx.QueryRow(ctx, insertUserFileStmt, queueID, task.UserID, task.FileName, task.SrcFileURL, task.SrcFileKey, task.DestFileURL, task.DestFileKey, "PENDING").Scan(&fileID, &createdAt)
		if err != nil {
			return nil, nil, fmt.Errorf("tx.QueryRow insertUserFileStmt %w", err)
		}

		queueItem := model.QueueItem{
			ID:    queueID,
			Order: orderNum,
		}
		queueItems = append(queueItems, queueItem)

		userItem := UserFileItem{
			ID:          fileID,
			QueueID:     queueID,
			UserID:      task.UserID,
			OrderNum:    orderNum,
			SrcFileURL:  task.SrcFileURL,
			SrcFileKey:  task.SrcFileKey,
			DestFileURL: task.DestFileURL,
			DestFileKey: task.DestFileKey,
			State:       "PENDING",
			CreatedAt:   createdAt,
			FileName:    task.FileName,
		}
		userItems = append(userItems, userItem)

		payload := cmodel.BrokerMessage{
			SrcFileURL:    task.SrcFileURL,
			DestFileURL:   task.DestFileURL,
			FileID:        fileID,
			FileExtension: filepath.Ext(task.FileName),
			DestFileKey:   task.DestFileKey,
			UserID:        task.UserID,
			QueueID:       queueID,
		}
		plJson, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("json.Marshal %w", err)
		}

		_, err = tx.Exec(ctx, insertOutboxStmt, plJson, strconv.FormatInt(fileID, 10))
		if err != nil {
			return nil, nil, fmt.Errorf("tx.Exec insertOutboxStmt %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("tx.Commit %w", err)
	}

	return queueItems, userItems, nil
}

// GetState retrieves the state of the user's tasks and files based on their user ID.
func (s *PostgresStore) GetState(ctx context.Context, userId int64) ([]model.ClientUserItem, error) {
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
		return nil, fmt.Errorf("error executing user state query: %w", err)
	}
	defer rows.Close()

	var userItems []model.ClientUserItem

	for rows.Next() {
		var item model.ClientUserItem
		err := rows.Scan(&item.Order, &item.FileName, &item.Link, &item.QueuePosition, &item.Status)
		if item.Status == "PENDING" {
			item.Link = ""
		}
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		userItems = append(userItems, item)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", rows.Err())
	}

	return userItems, nil
}

// CreateUser creates a new user in the database and returns the user ID.
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

// SendTasksToBroker retrieves tasks from the outbox and sends them to the message broker.
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

// FinishTasks updates the state of user files and removes the corresponding rows from the queue.
func (s *PostgresStore) FinishTasks(ctx context.Context, msgs []model.FinishedTask) error {
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
		err = tx.QueryRow(ctx, getQueueIDStmt, msg.FileID).Scan(&queueID)
		if err != nil {
			return fmt.Errorf("error retrieving queue_id: %w", err)
		}

		updateUserFilesStmt := `
			UPDATE userfiles
			SET state = $1, queue_id = NULL, dest_file_url = $3
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, updateUserFilesStmt, msg.Result, msg.FileID, msg.DestFileURL)
		if err != nil {
			return fmt.Errorf("error updating userfiles: %w", err)
		}

		deleteQueueStmt := `
			DELETE FROM queue
			WHERE id = $1
		`
		_, err = tx.Exec(ctx, deleteQueueStmt, queueID)
		if err != nil {
			return fmt.Errorf("error deleting from queue: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit %w", err)
	}

	return nil
}

// GetQueue retrieves all items from the queue.
func (s *PostgresStore) GetQueue(ctx context.Context) ([]model.QueueItem, error) {
	query := `SELECT id, order_num FROM queue ORDER BY order_num`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error executing queue query: %w", err)
	}
	defer rows.Close()

	var queueItems []model.QueueItem

	for rows.Next() {
		var item model.QueueItem
		if err := rows.Scan(&item.ID, &item.Order); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		queueItems = append(queueItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return queueItems, nil
}

// GetUserFiles retrieves all user files based on the user ID.
func (s *PostgresStore) GetUserFiles(ctx context.Context, userID int64) ([]UserFileItem, error) {
	query := `SELECT id, queue_id, user_id, order_num, src_file_url, src_file_key, dest_file_url, dest_file_key, state, created_at, file_name
		FROM userfiles
		WHERE user_id = $1
		ORDER BY order_num`
	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error executing userfiles query: %w", err)
	}
	defer rows.Close()

	var userFiles []UserFileItem

	for rows.Next() {
		var item UserFileItem
		var queueID sql.NullInt64

		if err := rows.Scan(&item.ID, &queueID, &item.UserID, &item.OrderNum, &item.SrcFileURL, &item.SrcFileKey, &item.DestFileURL, &item.DestFileKey, &item.State, &item.CreatedAt, &item.FileName); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		if queueID.Valid {
			item.QueueID = queueID.Int64
		} else {
			item.QueueID = 0
		}

		userFiles = append(userFiles, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return userFiles, nil
}

// Close closes the connection pool for the PostgresStore.
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}
