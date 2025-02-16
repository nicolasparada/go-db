package db

import (
	"context"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func executeTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	return crdb.ExecuteInTx(ctx, pgxTxAdapter{tx}, func() error {
		return fn(tx)
	})
}

type pgxTxAdapter struct {
	pgx.Tx
}

// Exec implements crdb.Tx interface.
func (a pgxTxAdapter) Exec(ctx context.Context, q string, args ...any) error {
	_, err := a.Tx.Exec(ctx, q, args...)
	return err
}
