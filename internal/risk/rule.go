package risk

import (
	"context"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
)

type Result struct {
	RuleName  string
	Score     float64
	Triggered bool
	Reason    string
}

type Rule interface {
	Name() string
	Evaluate(ctx context.Context, tx *domain.Transaction) (Result, error)
}
