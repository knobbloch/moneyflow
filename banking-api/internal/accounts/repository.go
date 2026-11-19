package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("account not found")
	ErrAlreadyBlocked = errors.New("account already blocked")
)

type Account struct {
	ID        string    `json:"id"`
	OwnerName string    `json:"owner_name"`
	Currency  string    `json:"currency"`
	Balance   int64     `json:"balance"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

const columns = `id::text, owner_name, currency, balance, status, created_at`

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scan(row pgx.Row) (*Account, error) {
	var a Account
	if err := row.Scan(&a.ID, &a.OwnerName, &a.Currency, &a.Balance, &a.Status, &a.CreatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) Create(ctx context.Context, owner, currency string, balance int64) (*Account, error) {
	return scan(r.pool.QueryRow(ctx,
		`INSERT INTO accounts (owner_name, currency, balance) VALUES ($1, $2, $3) RETURNING `+columns,
		owner, currency, balance))
}

func (r *Repository) Get(ctx context.Context, id string) (*Account, error) {
	a, err := scan(r.pool.QueryRow(ctx, `SELECT `+columns+` FROM accounts WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (r *Repository) Block(ctx context.Context, id string) (*Account, error) {
	a, err := scan(r.pool.QueryRow(ctx,
		`UPDATE accounts SET status = 'BLOCKED' WHERE id = $1 AND status = 'ACTIVE' RETURNING `+columns, id))
	if err == nil {
		return a, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)`, id).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	return nil, ErrAlreadyBlocked
}