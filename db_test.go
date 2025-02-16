package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
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
		assert.NoError(t, err)
		assert.Equal(t, 1, i)
	})

	t.Run("slice", func(t *testing.T) {
		var slice []int
		err := testDB.QueryRow(ctx, "SELECT ARRAY[1, 2, 3]").Scan(&slice)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, slice)
	})

	t.Run("json", func(t *testing.T) {
		type Data struct {
			A int `json:"a"`
		}

		var data Data
		err := testDB.QueryRow(ctx, `SELECT '{"a": 1}'::JSONB`).Scan(&data)
		assert.NoError(t, err)
		assert.Equal(t, Data{A: 1}, data)
	})

	t.Run("tx", func(t *testing.T) {
		_, err := testDB.Exec(ctx, "CREATE TABLE IF NOT EXISTS users (name VARCHAR PRIMARY KEY)")
		assert.NoError(t, err)

		err = testDB.RunTx(ctx, func(ctx context.Context) error {
			_, err := testDB.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "alice")
			assert.NoError(t, err)

			_, err = testDB.Exec(ctx, "UPDATE users SET name = $1 WHERE name = $2", "bob", "alice")
			assert.NoError(t, err)

			var name string
			err = testDB.QueryRow(ctx, "SELECT name FROM users WHERE name = $1", "alice").Scan(&name)
			assert.EqualError(t, err, "no rows in result set")

			err = testDB.QueryRow(ctx, "SELECT name FROM users WHERE name = $1", "bob").Scan(&name)
			assert.NoError(t, err)
			assert.Equal(t, "bob", name)

			return nil
		})
		assert.NoError(t, err)
	})
}
