package store

import (
	"context"

	"github.com/Alekseyt9/upscaler/internal/back/model"
)

type Store interface {
	// add to queue, userfiles, outbox
	CreateTasks(ctx context.Context, tasks []model.StoreTask) error

	// get user state
	GetState(ctx context.Context, userId int64) ([]model.UserItem, error)

	CreateUser(ctx context.Context) (int64, error)

	// chage userfiles state and delete queue row
	FinishTasks(ctx context.Context, ids []int64) error

	SendTasksToBroker(ctx context.Context, sendFunc func(items []model.OutboxItem) error) error

	Close() error
}
