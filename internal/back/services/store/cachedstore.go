package store

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/pkg/lrulom"
	"github.com/Alekseyt9/upscaler/pkg/ost"
)

type CachedStore struct {
	queue    *ost.OST
	queuemap map[int64]*CacheQueueItem
	lru      *lrulom.LRULoadOnMiss[int64, map[int64]*UserFileItem]
	dbstore  Store
	log      *slog.Logger
}

type CacheQueueItem struct {
	ID    int64
	Order int64
}

func (c CacheQueueItem) Less(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order < other.Order
}

func (c CacheQueueItem) Greater(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order > other.Order
}

func (c CacheQueueItem) Equal(item ost.Item) bool {
	other, ok := item.(CacheQueueItem)
	if !ok {
		return false
	}
	return c.Order == other.Order
}

func (c CacheQueueItem) Key() int {
	return int(c.ID)
}

func NewCachedStore(store Store, log *slog.Logger) (*CachedStore, error) {
	ctx := context.Background()
	lruLoadFunc := func(key int64) (map[int64]*UserFileItem, error) {
		v, err := store.GetUserFiles(ctx, key)
		m := make(map[int64]*UserFileItem, len(v))
		for _, item := range v {
			m[item.ID] = &item
		}
		return m, err
	}

	lru, err := lrulom.New(500, lruLoadFunc)
	if err != nil {
		return nil, err
	}

	cs := &CachedStore{
		queue:    ost.New(),
		queuemap: make(map[int64]*CacheQueueItem),
		lru:      lru,
		dbstore:  store,
		log:      log,
	}

	qs, err := store.GetQueue(ctx)
	if err != nil {
		return nil, err
	}

	for _, qi := range qs {
		cqi := CacheQueueItem{
			ID:    qi.ID,
			Order: qi.Order,
		}
		cs.queue.Insert(cqi)
		cs.queuemap[qi.ID] = &cqi
	}

	return cs, nil
}

func (s *CachedStore) CreateTasks(ctx context.Context, tasks []model.StoreTask) ([]model.QueueItem, []UserFileItem, error) {
	// first create in the database, then add to the cache
	qitems, filesitems, err := s.dbstore.CreateTasks(ctx, tasks)
	if err != nil {
		return nil, nil, err
	}

	for _, qi := range qitems {
		cqi := CacheQueueItem{
			ID:    qi.ID,
			Order: qi.Order,
		}
		s.queuemap[qi.ID] = &cqi
		s.queue.Insert(cqi)
	}

	for _, fi := range filesitems {
		ufiles, err := s.lru.GetOrLoad(fi.UserID)
		if err != nil {
			return nil, nil, err
		}
		ufiles[fi.ID] = &fi
	}

	return qitems, filesitems, nil
}

func (s *CachedStore) GetState(ctx context.Context, userId int64) ([]model.ClientUserItem, error) {
	items, err := s.lru.GetOrLoad(userId)
	if err != nil {
		return nil, fmt.Errorf("cached store lru.GetOrLoad %w", err)
	}

	res := make([]model.ClientUserItem, 0)
	for _, v := range items {
		var rank int
		qitem, ok := s.queuemap[v.QueueID]
		if ok {
			rank = s.queue.Rank(*qitem)
		}

		ci := model.ClientUserItem{
			Order:         v.OrderNum,
			FileName:      v.FileName,
			Link:          v.DestFileURL,
			QueuePosition: int64(rank),
			Status:        v.State,
		}
		res = append(res, ci)
	}

	slices.SortFunc(res, func(a, b model.ClientUserItem) int {
		return cmp.Compare(a.Order, b.Order)
	})

	return res, nil
}

func (s *CachedStore) FinishTasks(ctx context.Context, msgs []model.FinishedTask) error {
	err := s.dbstore.FinishTasks(ctx, msgs)
	if err != nil {
		return err
	}

	for _, m := range msgs {
		ufiles, err := s.lru.GetOrLoad(m.UserID)
		if err != nil {
			return err
		}

		qi, ok := s.queuemap[m.QueueID]
		if !ok {
			return fmt.Errorf("s.queuemap[m.QueueId]")
		}

		delete(s.queuemap, m.QueueID)
		s.queue.Delete(*qi)

		ufile, ok := ufiles[m.FileID]
		if ok {
			ufile.DestFileURL = m.DestFileURL
			ufile.State = m.Result
		}
	}

	return nil
}

func (s *CachedStore) CreateUser(ctx context.Context) (int64, error) {
	return s.dbstore.CreateUser(ctx)
}

func (s *CachedStore) SendTasksToBroker(ctx context.Context, sendFunc func(items []model.OutboxItem) error) error {
	return s.dbstore.SendTasksToBroker(ctx, sendFunc)
}

func (s *CachedStore) GetQueue(ctx context.Context) ([]model.QueueItem, error) {
	return s.dbstore.GetQueue(ctx)
}

func (s *CachedStore) GetUserFiles(ctx context.Context, userID int64) ([]UserFileItem, error) {
	return s.dbstore.GetUserFiles(ctx, userID)
}

func (s *CachedStore) Close() error {
	return s.dbstore.Close()
}
