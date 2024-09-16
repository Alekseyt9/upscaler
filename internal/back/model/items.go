package model

type StoreTask struct {
	UserID      int64  // ID of the user who owns the file
	SrcFileURL  string // URL of the source file
	SrcFileKey  string // Key of the source file
	DestFileURL string // URL of the destination (processed) file
	DestFileKey string // Key of the destination file
	//Payload     string // Payload to be inserted into the outbox table (must be a JSONB type)
}

type UserItem struct {
	Order         int
	FileName      string
	Link          string
	QueuePosition int
	Status        string
}

type BrokerMessage struct {
	SrcFileURL  string
	DestFileURL string
	TaskId      int64
}

type BrokerMessageResult struct {
	TaskId int64
	Result string
	Error  string
}
