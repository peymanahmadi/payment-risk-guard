package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/peymanahmadi/payment-risk-guard/internal/domain"
)

type AlertRepository struct {
	db *sql.DB
}

func NewAlertRepository(db *sql.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Save(ctx context.Context, a *domain.Alert) error {
	const q = `
		INSERT INTO alerts (id, transaction_id, account_id, score, severity, reasons, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, q,
		a.ID, a.TransactionID, a.AccountID, a.Score, string(a.Severity), pq.Array(a.Reasons), a.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert alert: %w", err)
	}
	return nil
}

func (r *AlertRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	const q = `
		SELECT id, transaction_id, account_id, score, severity, reasons, created_at
		FROM alerts WHERE id = $1`

	var a domain.Alert
	var severity string
	var alertID, txnID uuid.UUID
	err := r.db.QueryRowContext(ctx, q, id).Scan(&alertID, &txnID, &a.AccountID, &a.Score, &severity, pq.Array(&a.Reasons), &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query alert: %w", err)
	}
	a.ID = alertID
	a.TransactionID = txnID
	a.Severity = domain.Severity(severity)
	return &a, nil
}

func (r *AlertRepository) List(ctx context.Context, accountID string, minSeverity domain.Severity, limit, offset int) ([]domain.Alert, error) {
	q := `
		SELECT id, transaction_id, account_id, score, severity, reasons, created_at
		FROM alerts
		WHERE ($1 = '' OR account_id = $1)
		  AND ($2 = '' OR score >= (
		      CASE $2
		          WHEN 'low' THEN 0
		          WHEN 'medium' THEN 40
		          WHEN 'high' THEN 70
		          WHEN 'critical' THEN 90
		          ELSE 0
		      END
		  ))
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, q, accountID, string(minSeverity), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	var out []domain.Alert
	for rows.Next() {
		var a domain.Alert
		var severity string
		var alertID, txnID uuid.UUID
		if err := rows.Scan(&alertID, &txnID, &a.AccountID, &a.Score, &severity, pq.Array(&a.Reasons), &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan alert row: %w", err)
		}
		a.ID = alertID
		a.TransactionID = txnID
		a.Severity = domain.Severity(severity)
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate alert rows: %w", err)
	}
	return out, nil
}
