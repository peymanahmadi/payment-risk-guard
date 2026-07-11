package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
)

type AmountSpikeRule struct {
	Repo            domain.TransactionRepository
	Lookback        time.Duration
	SpikeMultiplier float64
	MinHistory      float64
}

func NewAmountSpikeRule(repo domain.TransactionRepository, lookback time.Duration, multiplier, minHistory float64) *AmountSpikeRule {
	return &AmountSpikeRule{Repo: repo, Lookback: lookback, SpikeMultiplier: multiplier, MinHistory: minHistory}
}

func (r *AmountSpikeRule) Name() string { return "amount_spike" }

func (r *AmountSpikeRule) Evaluate(ctx context.Context, tx *domain.Transaction) (Result, error) {
	avg, err := r.Repo.AverageAmountByAccount(ctx, tx.AccountID, r.Lookback)
	if err != nil {
		return Result{}, fmt.Errorf("amount spike rule: fetch average: %w", err)
	}

	if avg < r.MinHistory {
		return Result{RuleName: r.Name()}, nil
	}

	ratio := tx.Amount / avg
	if ratio < r.SpikeMultiplier {
		return Result{RuleName: r.Name()}, nil
	}

	score := clamp(30+ratio*8, 0, 100)

	return Result{
		RuleName:  r.Name(),
		Score:     score,
		Triggered: true,
		Reason:    fmt.Sprintf("amount %.2f is %.1fx the account's %s average (%.2f)", tx.Amount, ratio, r.Lookback, avg),
	}, nil
}
