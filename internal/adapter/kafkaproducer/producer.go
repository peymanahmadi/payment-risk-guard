package kafkaproducer

import (
	"context"
	"encoding/json"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type TransactionEvent struct {
	AccountID   string    `json:"account_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Country     string    `json:"country"`
	IPAddress   string    `json:"ip_address"`
	PaymentType string    `json:"payment_type"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type Producer struct {
	writer *kafka.Writer
}

func New(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.Hash{},
			AllowAutoTopicCreation: true,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, evt TransactionEvent) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.AccountID),
		Value: payload,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
