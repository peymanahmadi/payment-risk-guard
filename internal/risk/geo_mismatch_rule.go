package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
)

type GeoMismatchRule struct {
	Repo           domain.TransactionRepository
	LookbackWindow time.Duration
	MinGap         time.Duration
}

func NewGeoMismatchRule(repo domain.TransactionRepository, lookback, minGap time.Duration) *GeoMismatchRule {
	return &GeoMismatchRule{Repo: repo, LookbackWindow: lookback, MinGap: minGap}
}

func (r *GeoMismatchRule) Name() string { return "geo_mismatch" }

func (r *GeoMismatchRule) Evaluate(ctx context.Context, tx *domain.Transaction) (Result, error) {
	since := tx.OccurredAt.Add(-r.LookbackWindow)
	recent, err := r.Repo.RecentByAccount(ctx, tx.AccountID, since)
	if err != nil {
		return Result{}, fmt.Errorf("geo mismatch rule: fetch recent transactions: %w", err)
	}

	for _, prev := range recent {
		if prev.Country == "" || tx.Country == "" || prev.Country == tx.Country {
			continue
		}
		gap := tx.OccurredAt.Sub(prev.OccurredAt)
		if gap < 0 {
			gap = -gap
		}
		if gap < r.MinGap {
			score := clamp(60+(r.MinGap-gap).Minutes()/2, 0, 100)
			return Result{
				RuleName:  r.Name(),
				Score:     score,
				Triggered: true,
				Reason:    fmt.Sprintf("country changed %s -> %s within %s (minimum expected %s)", prev.Country, tx.Country, gap, r.MinGap),
			}, nil
		}
	}

	return Result{RuleName: r.Name()}, nil
}
