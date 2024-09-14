package model

import "time"

type StoreTask struct {
}

type UserItem struct {
	Order         time.Time
	FileName      string
	Link          string
	QueuePosition int
	Status        string
}
