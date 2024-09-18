package model

type BrokerMessage struct {
	SrcFileURL    string
	DestFileURL   string
	DestFileKey   string
	TaskId        int64
	FileExtension string
	UserID        string
}

type BrokerMessageResult struct {
	TaskId      int64
	Result      string
	Error       string
	DestFileKey string
	UserID      string
}

type BrokerOptions struct {
	Topic         string
	KafkaBrokers  []string
	ConsumerGroup string
}
