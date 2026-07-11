package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TransactionRepository interface {
	Save(ctx context.Context, tx *Transaction) error
	RecentByAccount(ctx context.Context, accountID string, since time.Time) ([]Transaction, error)
	AverageAmountByAccount(ctx context.Context, accountID string, lookback time.Duration) (float64, error)
	ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]Transaction, error)
}

type AlertRepository interface {
	Save(ctx context.Context, alert *Alert) error
	Get(ctx context.Context, id uuid.UUID) (*Alert, error)
	List(ctx context.Context, accountID string, minSeverity Severity, limit, offset int) ([]Alert, error)
}