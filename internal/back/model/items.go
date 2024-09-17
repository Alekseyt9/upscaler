package model

type StoreTask struct {
	UserID      int64 // ID of the user who owns the file
	FileName    string
	SrcFileURL  string // URL of the source file
	SrcFileKey  string // Key of the source file
	DestFileURL string // URL of the destination (processed) file
	DestFileKey string // Key of the destination file
}

type UploadedFile struct {
	Url  string
	Key  string
	Name string
}

type UserItem struct {
	Order         int
	FileName      string
	Link          string
	QueuePosition int
	Status        string
}

type OutboxItem struct {
	Payload string
	IdKey   string
}
