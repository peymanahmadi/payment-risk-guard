package kafkaconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
	"github.com/peymanahmadi/payment-risk-guard/internal/usecase"
)

type transactionEvent struct {
	AccountID   string    `json:"account_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Country     string    `json:"country"`
	IPAddress   string    `json:"ip_address"`
	PaymentType string    `json:"payment_type"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type Consumer struct {
	reader  *kafka.Reader
	process *usecase.ProcessTransaction
	log     *slog.Logger
}

func New(brokers []string, topic, groupID string, process *usecase.ProcessTransaction, log *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: reader, process: process, log: log}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.reader.Close()

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		if err := c.handle(ctx, msg); err != nil {
			c.log.Error("failed to process transaction event", "error", err, "offset", msg.Offset)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.Error("failed to commit offset", "error", err, "offset", msg.Offset)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, msg kafka.Message) error {
	var evt transactionEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return err
	}

	tx, err := domain.NewTransaction(evt.AccountID, evt.Amount, evt.Currency, evt.Country, evt.IPAddress, evt.PaymentType, evt.OccurredAt)
	if err != nil {
		return err
	}

	_, err = c.process.Execute(ctx, tx)
	return err
}
