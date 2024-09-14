package store

import "github.com/Alekseyt9/upscaler/internal/back/model"

type Store interface {
	// add to queue, userfiles, outbox
	CreateTask(model.StoreTask) error

	// get user state
	GetState(userId int64) ([]model.UserItem, error)
}
