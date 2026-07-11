package domain

import (
	"time"

	"github.com/google/uuid"
)

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

func SeverityFromScore(score float64) Severity {
	switch {
	case score >= 90:
		return SeverityCritical
	case score >= 70:
		return SeverityHigh
	case score >= 40:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

type Alert struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	AccountID     string    `json:"account_id"`
	Score         float64   `json:"score"`
	Severity      Severity  `json:"severity"`
	Reasons       []string  `json:"reasons"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewAlert(transactionID uuid.UUID, accountID string, score float64, reasons []string) *Alert {
	return &Alert{
		ID:            uuid.New(),
		TransactionID: transactionID,
		AccountID:     accountID,
		Score:         score,
		Severity:      SeverityFromScore(score),
		Reasons:       reasons,
		CreatedAt:     time.Now().UTC(),
	}
}
