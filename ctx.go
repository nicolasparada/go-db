package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var ctxKeyTx = struct{ name string }{name: "tx"}

func contextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, ctxKeyTx, tx)
}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(ctxKeyTx).(pgx.Tx)
	return tx, ok
}
