package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
)

type VelocityRule struct {
	Repo       domain.TransactionRepository
	Window     time.Duration
	MaxAllowed int
}

func NewVelocityRule(repo domain.TransactionRepository, window time.Duration, maxAllowed int) *VelocityRule {
	return &VelocityRule{Repo: repo, Window: window, MaxAllowed: maxAllowed}
}

func (r *VelocityRule) Name() string { return "velocity" }

func (r *VelocityRule) Evaluate(ctx context.Context, tx *domain.Transaction) (Result, error) {
	since := tx.OccurredAt.Add(-r.Window)
	recent, err := r.Repo.RecentByAccount(ctx, tx.AccountID, since)
	if err != nil {
		return Result{}, fmt.Errorf("velocity rule: fetch recent transactions: %w", err)
	}

	count := len(recent) + 1
	if count <= r.MaxAllowed {
		return Result{RuleName: r.Name()}, nil
	}

	over := float64(count - r.MaxAllowed)
	score := clamp(40+over*15, 0, 100)

	return Result{
		RuleName:  r.Name(),
		Score:     score,
		Triggered: true,
		Reason:    fmt.Sprintf("%d transactions in the last %s (limit %d)", count, r.Window, r.MaxAllowed),
	}, nil
}
