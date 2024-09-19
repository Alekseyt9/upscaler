package model

type StoreTask struct {
	UserID      int64 // ID of the user who owns the file
	FileName    string
	SrcFileURL  string // URL of the source file
	SrcFileKey  string // Key of the source file
	DestFileURL string // URL of the destination (processed) file
	DestFileKey string // Key of the destination file
}

type FinishedTask struct {
	FileID      int64
	UserID      int64
	QueueID     int64
	Result      string
	Error       string
	DestFileURL string
}

type UploadedFile struct {
	Url  string
	Key  string
	Name string
}

type ClientUserItem struct {
	Order         int64
	FileName      string
	Link          string
	QueuePosition int64
	Status        string
}

type OutboxItem struct {
	Payload string
	IdKey   string
}

type QueueItem struct {
	ID    int64
	Order int64
}
