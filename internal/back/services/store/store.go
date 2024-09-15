package store

import "github.com/Alekseyt9/upscaler/internal/back/model"

type Store interface {
	// add to queue, userfiles, outbox
	CreateTasks([]model.StoreTask) error

	// get user state
	GetState(userId int64) ([]model.UserItem, error)

	CreateUser() (int64, error)

	Close() error
}
