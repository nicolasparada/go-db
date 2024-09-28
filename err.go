package db

import (
	"cmp"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsError(err error, code string, cols ...string) bool {
	var e *pgconn.PgError
	if !errors.As(err, &e) {
		return false
	}

	if e.Code != code {
		return false
	}

	if len(cols) == 0 {
		return true
	}

	if e.ColumnName != "" {
		for _, col := range cols {
			if e.ColumnName == col {
				return true
			}
		}
	}

	if e.ConstraintName != "" {
		for _, col := range cols {
			if e.ConstraintName == col {
				return true
			}
		}
	}

	msg := strings.ToLower(e.Error())
	for _, col := range cols {
		if strings.Contains(msg, strings.ToLower(col)) {
			return true
		}
	}

	return false
}

func IsNotFoundError(err error) bool {
	return cmp.Or(
		errors.Is(err, pgx.ErrNoRows),
		errors.Is(err, sql.ErrNoRows),
	)
}

func IsNotNullViolationError(err error, cols ...string) bool {
	return IsError(err, pgerrcode.NotNullViolation, cols...)
}

func IsForeignKeyViolationError(err error, cols ...string) bool {
	return IsError(err, pgerrcode.ForeignKeyViolation, cols...)
}

func IsUniqueViolationError(err error, cols ...string) bool {
	return IsError(err, pgerrcode.UniqueViolation, cols...)
}
