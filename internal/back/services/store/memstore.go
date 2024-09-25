package store

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
)

// MemoryStore represents the in-memory store.
type MemoryStore struct {
	mu        sync.Mutex
	queue     []model.QueueItem
	userFiles map[int64][]UserFileItem
	users     map[int64]bool
	outbox    []model.OutboxItem
	idCounter int64
}

// NewMemoryStore initializes a new in-memory store with empty data structures.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		queue:     []model.QueueItem{},
		userFiles: make(map[int64][]UserFileItem),
		users:     make(map[int64]bool),
		outbox:    []model.OutboxItem{},
		idCounter: 1,
	}
}

// CreateTasks inserts tasks into the queue, user files, and outbox maps.
func (s *MemoryStore) CreateTasks(ctx context.Context, tasks []model.StoreTask) ([]model.QueueItem, []UserFileItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var queueItems []model.QueueItem
	var userItems []UserFileItem

	for _, task := range tasks {
		queueItem := model.QueueItem{
			ID:    s.idCounter,
			Order: s.idCounter,
		}
		s.queue = append(s.queue, queueItem)

		userFile := UserFileItem{
			ID:          s.idCounter,
			QueueID:     s.idCounter,
			UserID:      task.UserID,
			OrderNum:    s.idCounter,
			SrcFileURL:  task.SrcFileURL,
			SrcFileKey:  task.SrcFileKey,
			DestFileURL: task.DestFileURL,
			DestFileKey: task.DestFileKey,
			State:       "PENDING",
			CreatedAt:   time.Now(),
			FileName:    task.FileName,
		}
		s.userFiles[task.UserID] = append(s.userFiles[task.UserID], userFile)
		queueItems = append(queueItems, queueItem)
		userItems = append(userItems, userFile)

		payload := cmodel.BrokerMessage{
			SrcFileURL:    task.SrcFileURL,
			DestFileURL:   task.DestFileURL,
			FileID:        s.idCounter,
			FileExtension: filepath.Ext(task.FileName),
			DestFileKey:   task.DestFileKey,
			UserID:        task.UserID,
			QueueID:       s.idCounter,
		}
		plJson, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("json.Marshal %w", err)
		}

		outboxItem := model.OutboxItem{
			Payload: string(plJson),
			IdKey:   strconv.FormatInt(s.idCounter, 10),
		}
		s.outbox = append(s.outbox, outboxItem)

		s.idCounter++
	}

	return queueItems, userItems, nil
}

// GetState retrieves the state of the user's tasks and files based on their user ID.
func (s *MemoryStore) GetState(ctx context.Context, userID int64) ([]model.ClientUserItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userItems, ok := s.userFiles[userID]
	if !ok {
		return nil, nil
	}

	var clientUserItems []model.ClientUserItem

	for _, item := range userItems {
		clientItem := model.ClientUserItem{
			Order:         item.OrderNum,
			FileName:      item.FileName,
			Link:          item.DestFileURL,
			QueuePosition: -1,
			Status:        item.State,
		}

		if item.State == "PENDING" {
			clientItem.Link = ""
		}

		clientUserItems = append(clientUserItems, clientItem)
	}

	return clientUserItems, nil
}

// CreateUser creates a new user in the in-memory store and returns the user ID.
func (s *MemoryStore) CreateUser(ctx context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userID := s.idCounter
	s.users[userID] = true
	s.idCounter++

	return userID, nil
}

// SendTasksToBroker retrieves tasks from the outbox and sends them to the message broker.
func (s *MemoryStore) SendTasksToBroker(ctx context.Context, sendFunc func(items []model.OutboxItem) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.outbox) == 0 {
		return nil
	}

	var pendingItems []model.OutboxItem
	return sendFunc(pendingItems)
}

// FinishTasks updates the state of user files and removes corresponding rows from the queue.
func (s *MemoryStore) FinishTasks(ctx context.Context, msgs []model.FinishedTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, msg := range msgs {
		for i, userFile := range s.userFiles[msg.UserID] {
			if userFile.ID == msg.FileID {
				s.userFiles[msg.UserID][i].State = msg.Result
				s.userFiles[msg.UserID][i].DestFileURL = msg.DestFileURL
				s.userFiles[msg.UserID][i].QueueID = 0
				break
			}
		}

		for i, queueItem := range s.queue {
			if queueItem.ID == msg.QueueID {
				s.queue = append(s.queue[:i], s.queue[i+1:]...)
				break
			}
		}
	}

	return nil
}

// GetQueue retrieves all items from the queue.
func (s *MemoryStore) GetQueue(ctx context.Context) ([]model.QueueItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queue, nil
}

// GetUserFiles retrieves all user files based on the user ID.
func (s *MemoryStore) GetUserFiles(ctx context.Context, userID int64) ([]UserFileItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.userFiles[userID], nil
}

// Close is a no-op for MemoryStore, but provided for interface compatibility.
func (s *MemoryStore) Close() error {
	return nil
}
