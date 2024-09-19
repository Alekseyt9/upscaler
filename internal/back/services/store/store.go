package store

import (
	"context"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/model"
)

type UserFileItem struct {
	ID          int64
	QueueID     int64
	UserID      int64
	OrderNum    int64
	SrcFileURL  string
	SrcFileKey  string
	DestFileURL string
	DestFileKey string
	State       string
	CreatedAt   time.Time
	FileName    string
}

type Store interface {
	// add to queue, userfiles, outbox
	CreateTasks(ctx context.Context, tasks []model.StoreTask) ([]model.QueueItem, []UserFileItem, error)

	// get user state
	GetState(ctx context.Context, userId int64) ([]model.ClientUserItem, error)

	CreateUser(ctx context.Context) (int64, error)

	// chage userfiles state and delete queue row
	FinishTasks(ctx context.Context, msgs []model.FinishedTask) error

	SendTasksToBroker(ctx context.Context, sendFunc func(items []model.OutboxItem) error) error

	GetQueue(ctx context.Context) ([]model.QueueItem, error)

	GetUserFiles(ctx context.Context, userID int64) ([]UserFileItem, error)

	Close() error
}
