package db

import (
	"errors"
	"strings"

	"github.com/jackc/pgerrcode"
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

func IsNotNullViolationError(err error, cols ...string) bool {
	return IsError(err, pgerrcode.NotNullViolation, cols...)
}

func IsForeignKeyViolationError(err error, cols ...string) bool {
	return IsError(err, pgerrcode.ForeignKeyViolation, cols...)
}

func IsUniqueViolationError(err error, cols ...string) bool {
	return IsError(err, pgerrcode.UniqueViolation, cols...)
}
