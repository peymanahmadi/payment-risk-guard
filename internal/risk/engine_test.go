package risk_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
	"github.com/peymanahmadi/payment-risk-guard/internal/risk"
	"github.com/peymanahmadi/payment-risk-guard/internal/testutil"
)

type fakeRepo struct {
	mu   sync.Mutex
	txns []domain.Transaction
}

func newFakeRepo(seed ...domain.Transaction) *fakeRepo {
	return &fakeRepo{txns: seed}
}

func (f *fakeRepo) Save(_ context.Context, tx *domain.Transaction) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.txns = append(f.txns, *tx)
	return nil
}

func (f *fakeRepo) RecentByAccount(_ context.Context, accountID string, since time.Time) ([]domain.Transaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.Transaction
	for _, t := range f.txns {
		if t.AccountID == accountID && !t.OccurredAt.Before(since) {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeRepo) AverageAmountByAccount(_ context.Context, accountID string, lookback time.Duration) (float64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	since := time.Now().Add(-lookback)
	var sum float64
	var n int
	for _, t := range f.txns {
		if t.AccountID == accountID && t.OccurredAt.After(since) {
			sum += t.Amount
			n++
		}
	}
	if n == 0 {
		return 0, nil
	}
	return sum / float64(n), nil
}

func (f *fakeRepo) ListByAccount(_ context.Context, accountID string, limit, offset int) ([]domain.Transaction, error) {
	return f.RecentByAccount(context.Background(), accountID, time.Time{})
}

func mustTx(t *testing.T, accountID string, amount float64, country string, when time.Time) *domain.Transaction {
	t.Helper()
	tx, err := domain.NewTransaction(accountID, amount, "USD", country, "1.2.3.4", "deposit", when)
	testutil.NoError(t, err)
	return tx
}

func TestVelocityRule_TriggersOverLimit(t *testing.T) {
	now := time.Now()
	repo := newFakeRepo(
		*mustTx(t, "acc-1", 10, "US", now.Add(-1*time.Minute)),
		*mustTx(t, "acc-1", 10, "US", now.Add(-2*time.Minute)),
		*mustTx(t, "acc-1", 10, "US", now.Add(-3*time.Minute)),
	)
	rule := risk.NewVelocityRule(repo, 10*time.Minute, 3)

	res, err := rule.Evaluate(context.Background(), mustTx(t, "acc-1", 10, "US", now))
	testutil.NoError(t, err)
	testutil.True(t, res.Triggered, "velocity rule should trigger over the limit")
	testutil.Greater(t, res.Score, 0, "score should be positive")
}

func TestVelocityRule_DoesNotTriggerUnderLimit(t *testing.T) {
	now := time.Now()
	repo := newFakeRepo(*mustTx(t, "acc-1", 10, "US", now.Add(-1*time.Minute)))
	rule := risk.NewVelocityRule(repo, 10*time.Minute, 5)

	res, err := rule.Evaluate(context.Background(), mustTx(t, "acc-1", 10, "US", now))
	testutil.NoError(t, err)
	testutil.False(t, res.Triggered, "velocity rule should not trigger under the limit")
}
