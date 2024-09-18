package model

type BrokerMessage struct {
	SrcFileURL    string
	DestFileURL   string
	TaskId        int64
	FileExtension string
}

type BrokerMessageResult struct {
	TaskId int64
	Result string
	Error  string
}

type BrokerOptions struct {
	Topic         string
	KafkaBrokers  []string
	ConsumerGroup string
}
