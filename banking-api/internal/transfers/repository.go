package transfers

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound            = errors.New("transfer not found")
	ErrAccountNotFound     = errors.New("account not found")
	ErrAccountBlocked      = errors.New("account blocked")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
)

type Transfer struct {
	ID             string    `json:"id"`
	FromAccountID  string    `json:"from_account_id"`
	ToAccountID    string    `json:"to_account_id"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type Input struct {
	From   string
	To     string
	Amount int64
	Key    string
}

const columns = `id::text, from_account_id::text, to_account_id::text, amount, currency, status, idempotency_key, created_at`

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scan(row pgx.Row) (*Transfer, error) {
	var t Transfer
	err := row.Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Currency, &t.Status, &t.IdempotencyKey, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type lockedAccount struct {
	currency string
	balance  int64
	status   string
}

func (r *Repository) Create(ctx context.Context, in Input) (*Transfer, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, in.Key); err != nil {
		return nil, false, err
	}

	existing, err := scan(tx.QueryRow(ctx, `SELECT `+columns+` FROM transfers WHERE idempotency_key = $1`, in.Key))
	if err == nil {
		if existing.FromAccountID == in.From && existing.ToAccountID == in.To && existing.Amount == in.Amount {
			return existing, true, nil
		}
		return nil, false, ErrIdempotencyConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	ids := []string{in.From, in.To}
	sort.Strings(ids)
	locked := make(map[string]lockedAccount, 2)
	for _, id := range ids {
		var a lockedAccount
		err := tx.QueryRow(ctx,
			`SELECT currency, balance, status FROM accounts WHERE id = $1 FOR UPDATE`, id,
		).Scan(&a.currency, &a.balance, &a.status)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, ErrAccountNotFound
		}
		if err != nil {
			return nil, false, err
		}
		locked[id] = a
	}

	from, to := locked[in.From], locked[in.To]
	if from.status != "ACTIVE" || to.status != "ACTIVE" {
		return nil, false, ErrAccountBlocked
	}
	if from.currency != to.currency {
		return nil, false, ErrCurrencyMismatch
	}
	if from.balance < in.Amount {
		return nil, false, ErrInsufficientFunds
	}

	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance - $1 WHERE id = $2`, in.Amount, in.From); err != nil {
		return nil, false, err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, in.Amount, in.To); err != nil {
		return nil, false, err
	}

	t, err := scan(tx.QueryRow(ctx,
		`INSERT INTO transfers (from_account_id, to_account_id, amount, currency, idempotency_key)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+columns,
		in.From, in.To, in.Amount, from.currency, in.Key))
	if err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return t, false, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*Transfer, error) {
	t, err := scan(r.pool.QueryRow(ctx, `SELECT `+columns+` FROM transfers WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *Repository) ListByAccount(ctx context.Context, accountID string, limit, offset int) ([]Transfer, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)`, accountID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAccountNotFound
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+columns+` FROM transfers
		 WHERE from_account_id = $1 OR to_account_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Transfer{}
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *t)
	}
	return items, rows.Err()
}