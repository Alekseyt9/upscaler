package cache

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/pkg/lrulom"
	"github.com/Alekseyt9/upscaler/pkg/ost"
)

// CachedStore is a caching layer that wraps around a store and uses an in-memory cache
type CachedStore struct {
	queue    *ost.POST                                                   // OST (Order Statistics Tree) for managing the order of tasks in the queue.
	queuemap sync.Map                                                    // Map to store queue items for quick lookup.
	lru      *lrulom.LRULoadOnMiss[int64, map[int64]*store.UserFileItem] // LRU cache that loads user files on miss from the database.
	dbstore  store.Store                                                 // Underlying database store.
	log      *slog.Logger                                                // Logger for logging operations and errors.
}

// NewCachedStore initializes a new CachedStore by wrapping a database store and loading the initial queue state into memory.
func NewCachedStore(s store.Store, log *slog.Logger) (*CachedStore, error) {
	ctx := context.Background()

	// Function to load user files into the cache when they are not already present.
	lruLoadFunc := func(key int64) (map[int64]*store.UserFileItem, error) {
		v, err := s.GetUserFiles(ctx, key)
		m := make(map[int64]*store.UserFileItem, len(v))
		for _, item := range v {
			m[item.ID] = &item
		}
		return m, err
	}

	lru, err := lrulom.New(500, lruLoadFunc) // Initialize LRU cache with a capacity of 500 entries.
	if err != nil {
		return nil, err
	}

	cs := &CachedStore{
		queue:    ost.NewPOST(),
		queuemap: sync.Map{},
		lru:      lru,
		dbstore:  s,
		log:      log,
	}

	// Load the current queue state from the database and populate the cache.
	qs, err := s.GetQueue(ctx)
	if err != nil {
		return nil, err
	}

	// Insert queue items into both the OST and queuemap.
	for _, qi := range qs {
		cqi := CacheQueueItem{
			ID:    qi.ID,
			Order: qi.Order,
		}
		cs.queue.Insert(cqi)
		cs.queuemap.Store(qi.ID, &cqi)
	}

	return cs, nil
}

// CreateTasks creates new tasks by first adding them to the database store,
// then adding them to the cache.
func (s *CachedStore) CreateTasks(ctx context.Context, tasks []model.StoreTask) ([]model.QueueItem, []store.UserFileItem, error) {
	// Create tasks in the database.
	qitems, filesitems, err := s.dbstore.CreateTasks(ctx, tasks)
	if err != nil {
		return nil, nil, err
	}

	// Add new queue items to the in-memory queue.
	for _, qi := range qitems {
		cqi := CacheQueueItem{
			ID:    qi.ID,
			Order: qi.Order,
		}
		s.queuemap.Store(qi.ID, &cqi)
		s.queue.Insert(cqi)
	}

	// Update the LRU cache with new file items.
	for _, fi := range filesitems {
		ufiles, err := s.lru.GetOrLoad(fi.UserID)
		if err != nil {
			return nil, nil, err
		}
		ufiles[fi.ID] = &fi
	}

	return qitems, filesitems, nil
}

// GetState retrieves the user's task state from the cache.
// It looks up the user's tasks in the LRU cache and calculates their position in the queue.
func (s *CachedStore) GetState(ctx context.Context, userId int64) ([]model.ClientUserItem, error) {
	items, err := s.lru.GetOrLoad(userId)
	if err != nil {
		return nil, fmt.Errorf("cached store lru.GetOrLoad %w", err)
	}

	// Prepare the result set for the user's state.
	res := make([]model.ClientUserItem, 0)
	for _, v := range items {
		var rank int
		value, ok := s.queuemap.Load(v.QueueID)
		if ok {
			qitem := value.(*CacheQueueItem)
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

// FinishTasks updates the task state by first updating the database store,
// then updating the cache by removing the completed tasks.
func (s *CachedStore) FinishTasks(ctx context.Context, msgs []model.FinishedTask) error {
	err := s.dbstore.FinishTasks(ctx, msgs)
	if err != nil {
		return err
	}

	// Update the cache by removing finished tasks from both the queue and LRU cache.
	for _, m := range msgs {
		ufiles, err := s.lru.GetOrLoad(m.UserID)
		if err != nil {
			return err
		}

		value, ok := s.queuemap.Load(m.QueueID)
		if !ok {
			return fmt.Errorf("s.queuemap[m.QueueId]")
		}

		qi := value.(*CacheQueueItem)
		s.queuemap.Delete(m.QueueID)
		s.queue.Delete(*qi)

		ufile, ok := ufiles[m.FileID]
		if ok {
			ufile.DestFileURL = m.DestFileURL
			ufile.State = m.Result
		}
	}

	return nil
}

// CreateUser delegates the creation of a user to the database store.
func (s *CachedStore) CreateUser(ctx context.Context) (int64, error) {
	return s.dbstore.CreateUser(ctx)
}

// SendTasksToBroker delegates sending tasks to the broker to the database store.
func (s *CachedStore) SendTasksToBroker(ctx context.Context, sendFunc func(items []model.OutboxItem) error) error {
	return s.dbstore.SendTasksToBroker(ctx, sendFunc)
}

// GetQueue delegates the retrieval of the task queue to the database store.
func (s *CachedStore) GetQueue(ctx context.Context) ([]model.QueueItem, error) {
	return s.dbstore.GetQueue(ctx)
}

// GetUserFiles retrieves the user's files from the database store.
func (s *CachedStore) GetUserFiles(ctx context.Context, userID int64) ([]store.UserFileItem, error) {
	return s.dbstore.GetUserFiles(ctx, userID)
}

// Close closes the underlying database store.
func (s *CachedStore) Close() error {
	return s.dbstore.Close()
}
