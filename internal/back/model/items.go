// Package model defines various data structures used throughout the application.
package model

// StoreTask represents a task stored in the system for processing a file.
type StoreTask struct {
	UserID      int64  // ID of the user who owns the file
	FileName    string // Name of the file
	SrcFileURL  string // URL of the source file
	SrcFileKey  string // Key of the source file in storage
	DestFileURL string // URL of the destination (processed) file
	DestFileKey string // Key of the destination file in storage
}

// FinishedTask represents a task that has been processed and completed.
type FinishedTask struct {
	FileID      int64  // ID of the file
	UserID      int64  // ID of the user who owns the file
	QueueID     int64  // ID of the queue item
	Result      string // Result of the processing (e.g., "PROCESSED", "ERROR")
	Error       string // Error message if processing failed
	DestFileURL string // URL of the destination (processed) file
}

// UploadedFile represents a file that has been uploaded to the system.
type UploadedFile struct {
	Url  string // URL of the uploaded file
	Key  string // Key of the uploaded file in storage
	Name string // Name of the uploaded file
}

// ClientUserItem represents an item in the client's queue or list.
type ClientUserItem struct {
	Order         int64  // Order of the item in the list
	FileName      string // Name of the file
	Link          string // Download link for the file
	QueuePosition int64  // Position of the file in the queue
	Status        string // Status of the file (e.g., "PENDING", "PROCESSED", "ERROR")
}

// OutboxItem represents an item in the outbox, which is used for reliable message processing.
type OutboxItem struct {
	Payload string // The message payload
	IdKey   string // Unique identifier key for the message
}

// QueueItem represents an item in the processing queue.
type QueueItem struct {
	ID    int64 // ID of the queue item
	Order int64 // Order of the item in the queue
}
