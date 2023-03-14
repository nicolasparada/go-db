package db

import (
	"context"

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
	}
	return db.pool.Query(ctx, query, args...)
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if tx, ok := txFromContext(ctx); ok {
		return tx.QueryRow(ctx, query, args...)
	}
	return db.pool.QueryRow(ctx, query, args...)
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if tx, ok := txFromContext(ctx); ok {
		return tx.Exec(ctx, query, args...)
	}
	return db.pool.Exec(ctx, query, args...)
}

func (db *DB) RunTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}

	return executeTx(ctx, db.pool, func(tx pgx.Tx) error {
		return fn(contextWithTx(ctx, tx))
	})
}
