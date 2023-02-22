package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

func (db *DB) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Query(ctx, query, args...)
	} else {
		return db.pool.Query(ctx, query, args...)
	}
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if tx, ok := txFromContext(ctx); ok {
		return tx.QueryRow(ctx, query, args...)
	} else {
		return db.pool.QueryRow(ctx, query, args...)
	}
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Exec(ctx, query, args...)
	} else {
		return db.pool.Exec(ctx, query, args...)
	}
}

func (db *DB) RunTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, errTx := db.pool.Begin(ctx)
	if err != nil {
		return errTx
	}

	defer func() {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil && !errors.Is(err, pgx.ErrTxClosed) {
			err = errRollback
		}
	}()

	errFunc := fn(contextWithTx(ctx, tx))
	if errFunc != nil {
		_ = tx.Rollback(ctx)
		return errFunc
	}

	return tx.Commit(ctx)
}
