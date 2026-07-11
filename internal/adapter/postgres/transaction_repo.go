package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Save(ctx context.Context, tx *domain.Transaction) error {
	const q = `
		INSERT INTO transactions (id, account_id, amount, currency, country, ip_address, payment_type, occurred_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, q,
		tx.ID, tx.AccountID, tx.Amount, tx.Currency, tx.Country, tx.IPAddress, tx.PaymentType, tx.OccurredAt, tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) RecentByAccount(ctx context.Context, accountID string, since time.Time) ([]domain.Transaction, error) {
	const q = `
		SELECT id, account_id, amount, currency, country, ip_address, payment_type, occurred_at, created_at
		FROM transactions
		WHERE account_id = $1 AND occurred_at >= $2
		ORDER BY occurred_at DESC`

	rows, err := r.db.QueryContext(ctx, q, accountID, since)
	if err != nil {
		return nil, fmt.Errorf("query recent transactions: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

func (r *TransactionRepository) AverageAmountByAccount(ctx context.Context, accountID string, lookback time.Duration) (float64, error) {
	const q = `
		SELECT COALESCE(AVG(amount), 0)
		FROM transactions
		WHERE account_id = $1 AND occurred_at >= $2`

	var avg float64
	since := time.Now().UTC().Add(-lookback)
	if err := r.db.QueryRowContext(ctx, q, accountID, since).Scan(&avg); err != nil {
		return 0, fmt.Errorf("query average amount: %w", err)
	}
	return avg, nil
}

func (r *TransactionRepository) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]domain.Transaction, error) {
	const q = `
		SELECT id, account_id, amount, currency, country, ip_address, payment_type, occurred_at, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, q, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query transactions by account: %w", err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

func scanTransactions(rows *sql.Rows) ([]domain.Transaction, error) {
	var out []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		var id uuid.UUID
		if err := rows.Scan(&id, &t.AccountID, &t.Amount, &t.Currency, &t.Country, &t.IPAddress, &t.PaymentType, &t.OccurredAt, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		t.ID = id
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transaction rows: %w", err)
	}
	return out, nil
}
