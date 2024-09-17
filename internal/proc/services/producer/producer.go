package producer

import "github.com/Alekseyt9/upscaler/internal/common/model"

type Producer struct {
}

func NewProducer() *Producer {
	return nil
}

func (b *Producer) Send(msg model.BrokerMessageResult) error {
	return nil
}
