package store

import (
	"context"
	"time"

	"github.com/Alekseyt9/upscaler/internal/back/model"
)

// UserFileItem represents a file item associated with a user in the system.
type UserFileItem struct {
	ID          int64     // Unique identifier for the file item.
	QueueID     int64     // ID of the associated queue item.
	UserID      int64     // ID of the user who owns the file.
	OrderNum    int64     // The order number of the file in the processing queue.
	SrcFileURL  string    // URL for downloading the source file.
	SrcFileKey  string    // Key for the source file in storage.
	DestFileURL string    // URL for downloading the processed destination file.
	DestFileKey string    // Key for the destination file in storage.
	State       string    // Current state of the file processing (e.g., "PENDING", "PROCESSED").
	CreatedAt   time.Time // Timestamp when the file was created.
	FileName    string    // The name of the file.
}

// Store defines the interface for interacting with the application's data store.
// It provides methods for creating tasks, managing user state, finishing tasks,
// and interacting with the task queue and outbox.
type Store interface {
	// CreateTasks adds new tasks to the queue, userfiles, and outbox tables.
	//
	// Parameters:
	//   - ctx: Context for managing request deadlines and cancellation signals.
	//   - tasks: A slice of StoreTask items representing tasks to be created.
	//
	// Returns:
	//   - A slice of QueueItem items representing the tasks in the processing queue.
	//   - A slice of UserFileItem items representing the user files associated with the tasks.
	//   - An error if there is a failure in creating the tasks.
	CreateTasks(ctx context.Context, tasks []model.StoreTask) ([]model.QueueItem, []UserFileItem, error)

	// GetState retrieves the current state of the user's files.
	//
	// Parameters:
	//   - ctx: Context for managing request deadlines and cancellation signals.
	//   - userId: The ID of the user whose file state is being retrieved.
	//
	// Returns:
	//   - A slice of ClientUserItem items representing the current state of the user's files.
	//   - An error if there is a failure in retrieving the user's file state.
	GetState(ctx context.Context, userId int64) ([]model.ClientUserItem, error)

	// CreateUser creates a new user in the system.
	//
	// Parameters:
	//   - ctx: Context for managing request deadlines and cancellation signals.
	//
	// Returns:
	//   - The ID of the newly created user.
	//   - An error if there is a failure in creating the user.
	CreateUser(ctx context.Context) (int64, error)

	// FinishTasks updates the state of user files and removes the associated tasks from the queue.
	//
	// Parameters:
	//   - ctx: Context for managing request deadlines and cancellation signals.
	//   - msgs: A slice of FinishedTask items representing tasks that have been completed.
	//
	// Returns:
	//   - An error if there is a failure in finishing the tasks.
	FinishTasks(ctx context.Context, msgs []model.FinishedTask) error

	// SendTasksToBroker sends tasks from the outbox to the message broker.
	//
	// Parameters:
	//   - ctx: Context for managing request deadlines and cancellation signals.
	//   - sendFunc: A function that processes and sends the outbox items to the message broker.
	//
	// Returns:
	//   - An error if there is a failure in sending tasks to the broker.
	SendTasksToBroker(ctx context.Context, sendFunc func(items []model.OutboxItem) error) error

	// GetQueue retrieves the current processing queue.
	//
	// Parameters:
	//   - ctx: Context for managing request deadlines and cancellation signals.
	//
	// Returns:
	//   - A slice of QueueItem items representing the tasks in the processing queue.
	//   - An error if there is a failure in retrieving the queue.
	GetQueue(ctx context.Context) ([]model.QueueItem, error)

	// GetUserFiles retrieves the files associated with a user.
	//
	// Parameters:
	//   - ctx: Context for managing request deadlines and cancellation signals.
	//   - userID: The ID of the user whose files are being retrieved.
	//
	// Returns:
	//   - A slice of UserFileItem items representing the user's files.
	//   - An error if there is a failure in retrieving the user's files.
	GetUserFiles(ctx context.Context, userID int64) ([]UserFileItem, error)

	// Close gracefully shuts down the store and releases any resources.
	//
	// Returns:
	//   - An error if there is a failure in closing the store.
	Close() error
}
