package usecase_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
	"github.com/peymanahmadi/payment-risk-guard/internal/risk"
	"github.com/peymanahmadi/payment-risk-guard/internal/testutil"
	"github.com/peymanahmadi/payment-risk-guard/internal/usecase"
)

type memTxRepo struct {
	mu   sync.Mutex
	txns []domain.Transaction
}

func (r *memTxRepo) Save(_ context.Context, tx *domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.txns = append(r.txns, *tx)
	return nil
}

func (r *memTxRepo) RecentByAccount(_ context.Context, accountID string, since time.Time) ([]domain.Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.Transaction
	for _, t := range r.txns {
		if t.AccountID == accountID && !t.OccurredAt.Before(since) {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *memTxRepo) AverageAmountByAccount(_ context.Context, accountID string, lookback time.Duration) (float64, error) {
	return 0, nil
}

func (r *memTxRepo) ListByAccount(_ context.Context, accountID string, limit, offset int) ([]domain.Transaction, error) {
	return r.RecentByAccount(context.Background(), accountID, time.Time{})
}

type memAlertRepo struct {
	mu     sync.Mutex
	alerts []domain.Alert
}

func (r *memAlertRepo) Save(_ context.Context, a *domain.Alert) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.alerts = append(r.alerts, *a)
	return nil
}

func (r *memAlertRepo) Get(_ context.Context, id uuid.UUID) (*domain.Alert, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.alerts {
		if a.ID == id {
			cp := a
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *memAlertRepo) List(_ context.Context, accountID string, minSeverity domain.Severity, limit, offset int) ([]domain.Alert, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.Alert
	for _, a := range r.alerts {
		if accountID == "" || a.AccountID == accountID {
			out = append(out, a)
		}
	}
	return out, nil
}

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

func TestProcessTransaction_RaisesAlertWhenRiskCrossesThreshold(t *testing.T) {
	txRepo := &memTxRepo{}
	alertRepo := &memAlertRepo{}
	engine := risk.NewEngine(50, risk.NewVelocityRule(txRepo, 10*time.Minute, 1))
	uc := usecase.NewProcessTransaction(txRepo, alertRepo, engine, noopLogger{})

	now := time.Now()
	first, err := domain.NewTransaction("acc-1", 10, "USD", "US", "1.1.1.1", "deposit", now.Add(-time.Minute))
	testutil.NoError(t, err)
	_, err = uc.Execute(context.Background(), first)
	testutil.NoError(t, err)

	second, err := domain.NewTransaction("acc-1", 10, "USD", "US", "1.1.1.1", "deposit", now)
	testutil.NoError(t, err)
	outcome, err := uc.Execute(context.Background(), second)
	testutil.NoError(t, err)

	testutil.True(t, outcome.Verdict.Triggered, "verdict should be triggered")
	if outcome.Alert == nil {
		t.Fatal("expected an alert to be raised")
	}
	testutil.LenInt(t, len(alertRepo.alerts), 1, "one alert should be saved")
	testutil.LenInt(t, len(txRepo.txns), 2, "both transactions should be saved")
}
