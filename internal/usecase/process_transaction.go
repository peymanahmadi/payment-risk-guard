package usecase

import (
	"context"
	"fmt"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
	"github.com/peymanahmadi/payment-risk-guard/internal/risk"
)

type Logger interface {
	Info(msg string, kv ...any)
	Error(msg string, kv ...any)
}

type ProcessTransaction struct {
	Transactions domain.TransactionRepository
	Alerts       domain.AlertRepository
	Engine       *risk.Engine
	Log          Logger
}

func NewProcessTransaction(txRepo domain.TransactionRepository, alertRepo domain.AlertRepository, engine *risk.Engine, log Logger) *ProcessTransaction {
	return &ProcessTransaction{Transactions: txRepo, Alerts: alertRepo, Engine: engine, Log: log}
}

type Outcome struct {
	Transaction *domain.Transaction
	Verdict     risk.Verdict
	Alert       *domain.Alert
}

func (uc *ProcessTransaction) Execute(ctx context.Context, tx *domain.Transaction) (Outcome, error) {
	verdict, err := uc.Engine.Evaluate(ctx, tx)
	if err != nil {
		return Outcome{}, fmt.Errorf("evaluate risk: %w", err)
	}

	if err := uc.Transactions.Save(ctx, tx); err != nil {
		return Outcome{}, fmt.Errorf("save transaction: %w", err)
	}

	outcome := Outcome{Transaction: tx, Verdict: verdict}

	if verdict.Triggered {
		alert := domain.NewAlert(tx.ID, tx.AccountID, verdict.Score, verdict.Reasons)
		if err := uc.Alerts.Save(ctx, alert); err != nil {
			return Outcome{}, fmt.Errorf("save alert: %w", err)
		}
		outcome.Alert = alert
		uc.Log.Info("risk alert raised", "account_id", tx.AccountID, "transaction_id", tx.ID, "score", verdict.Score, "severity", alert.Severity)
	}

	return outcome, nil
}
