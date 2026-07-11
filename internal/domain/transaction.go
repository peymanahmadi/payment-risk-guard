package domain

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID uuid.UUID `json:"id"`
	AccountID string `json:"account_id"`
	Amount float64 `json:"amount"`
	Currency string `json:"currency"`
	Country string `json:"country"`
	IPAddress string `json:"ip_address"`
	PaymentType string `json:"payment_type"`
	OccurredAt time.Time `json:"occurred_at"`
	CreatedAt time.Time `json:"created_at"`
}

func NewTransaction(accountID string, amount float64, currency, country, ip, paymentType string, occurredAt time.Time) (*Transaction, error) {
	if accountID == "" {
		return nil, ErrInvalidAccountID
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	return &Transaction{
		ID: uuid.New(),
		AccountID: accountID,
		Amount: amount,
		Currency: currency,
		Country: country,
		IPAddress: ip,
		PaymentType: paymentType,
		OccurredAt: occurredAt.UTC(),
		CreatedAt: time.Now().UTC(),
	}, nil
}