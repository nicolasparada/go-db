package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
)

var testDB *DB

func TestMain(m *testing.M) {
	code, err := setupT(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "setupT() failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func setupT(m *testing.M) (int, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return 0, err
	}

	if err := pool.Client.Ping(); err != nil {
		return 0, err
	}

	cockroach, err := setupCockroach(pool)
	if err != nil {
		return 0, err
	}

	defer cockroach.Close()

	dbPool, err := setupDB(cockroach, pool.Retry)
	if err != nil {
		return 0, err
	}

	defer dbPool.Close()

	testDB = &DB{pool: dbPool}

	return m.Run(), nil
}

func setupCockroach(pool *dockertest.Pool) (*dockertest.Resource, error) {
	return pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "cockroachdb/cockroach",
		Tag:        "latest",
		Cmd: []string{"start-single-node",
			"--insecure",
			"--store", "type=mem,size=0.25",
			"--advertise-addr", "localhost",
		},
	})
}

func setupDB(cockroach *dockertest.Resource, retry func(op func() error) error) (*pgxpool.Pool, error) {
	ctx := context.Background()
	var pool *pgxpool.Pool
	return pool, retry(func() (err error) {
		hostPort := cockroach.GetHostPort("26257/tcp")
		pool, err = pgxpool.New(ctx, "postgresql://root@"+hostPort+"/defaultdb?sslmode=disable")
		if err != nil {
			return err
		}

		return pool.Ping(ctx)
	})
}

func TestDB_QueryRow(t *testing.T) {
	ctx := t.Context()
	t.Run("int", func(t *testing.T) {
		var i int
		err := testDB.QueryRow(ctx, "SELECT 1").Scan(&i)
		assertNil(t, err)
		assertEqual(t, i, 1)
	})

	t.Run("slice", func(t *testing.T) {
		var slice []int
		err := testDB.QueryRow(ctx, "SELECT ARRAY[1, 2, 3]").Scan(&slice)
		assertNil(t, err)
		assertEqual(t, slice, []int{1, 2, 3})
	})

	t.Run("json", func(t *testing.T) {
		type Data struct {
			A int `json:"a"`
		}

		var data Data
		err := testDB.QueryRow(ctx, `SELECT '{"a": 1}'::JSONB`).Scan(&data)
		assertNil(t, err)
		assertEqual(t, data, Data{A: 1})
	})

	t.Run("tx", func(t *testing.T) {
		_, err := testDB.Exec(ctx, "CREATE TABLE IF NOT EXISTS users (name VARCHAR PRIMARY KEY)")
		assertNil(t, err)

		err = testDB.RunTx(ctx, func(ctx context.Context) error {
			_, err := testDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "alice")
			assertNil(t, err)

			_, err = testDB.Exec(ctx, "UPDATE users SET name = $1 WHERE name = $2", "bob", "alice")
			assertNil(t, err)

			var name string
			err = testDB.QueryRow(ctx, "SELECT name FROM users WHERE name = $1", "alice").Scan(&name)
			assertErrorIs(t, err, sql.ErrNoRows)

			err = testDB.QueryRow(ctx, "SELECT name FROM users WHERE name = $1", "bob").Scan(&name)
			assertNil(t, err)
			assertEqual(t, name, "bob")

			return nil
		})
		assertNil(t, err)
	})
}

func assertEqual[T any](t *testing.T, got, want T) {
	t.Helper()
	if !isEqual(got, want) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func assertNil(t *testing.T, got any) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("got: %v; want: nil", got)
	}
}

func assertErrorIs(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func isEqual[T any](got, want T) bool {
	if isNil(got) && isNil(want) {
		return true
	}
	if equalable, ok := any(got).(interface{ Equal(T) bool }); ok {
		return equalable.Equal(want)
	}
	return reflect.DeepEqual(got, want)
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}
	return false
}
