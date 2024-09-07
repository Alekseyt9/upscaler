package store

type Store interface {
	SaveMessage()
	LoadMessages()
	MarkSendedMessages()

	SaveTask()
}
