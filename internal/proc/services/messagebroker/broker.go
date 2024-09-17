package messagebroker

import "github.com/Alekseyt9/upscaler/internal/common/model"

type MessageBroker struct {
}

func (b *MessageBroker) Send(msg model.BrokerMessageResult) error {
	return nil
}
