package httpapi

import (
	"time"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
	"github.com/peymanahmadi/payment-risk-guard/internal/usecase"
)

type createTransactionRequest struct {
	AccountID   string    `json:"account_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Country     string    `json:"country"`
	IPAddress   string    `json:"ip_address"`
	PaymentType string    `json:"payment_type"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type transactionResponse struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Country     string    `json:"country"`
	PaymentType string    `json:"payment_type"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type alertResponse struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	AccountID     string    `json:"account_id"`
	Score         float64   `json:"score"`
	Severity      string    `json:"severity"`
	Reasons       []string  `json:"reasons"`
	CreatedAt     time.Time `json:"created_at"`
}

type processTransactionResponse struct {
	Transaction transactionResponse `json:"transaction"`
	RiskScore   float64             `json:"risk_score"`
	Flagged     bool                `json:"flagged"`
	Reasons     []string            `json:"reasons"`
	Alert       *alertResponse      `json:"alert,omitempty"`
}

func toTransactionResponse(t *domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:          t.ID.String(),
		AccountID:   t.AccountID,
		Amount:      t.Amount,
		Currency:    t.Currency,
		Country:     t.Country,
		PaymentType: t.PaymentType,
		OccurredAt:  t.OccurredAt,
	}
}

func toAlertResponse(a *domain.Alert) *alertResponse {
	if a == nil {
		return nil
	}
	return &alertResponse{
		ID:            a.ID.String(),
		TransactionID: a.TransactionID.String(),
		AccountID:     a.AccountID,
		Score:         a.Score,
		Severity:      string(a.Severity),
		Reasons:       a.Reasons,
		CreatedAt:     a.CreatedAt,
	}
}

func toOutcomeResponse(o usecase.Outcome) processTransactionResponse {
	return processTransactionResponse{
		Transaction: toTransactionResponse(o.Transaction),
		RiskScore:   o.Verdict.Score,
		Flagged:     o.Verdict.Triggered,
		Reasons:     o.Verdict.Reasons,
		Alert:       toAlertResponse(o.Alert),
	}
}
