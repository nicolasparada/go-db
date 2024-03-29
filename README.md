# Golang Database Wrapper

[![Go Reference](https://pkg.go.dev/badge/github.com/nicolasparada/go-db.svg)](https://pkg.go.dev/github.com/nicolasparada/go-db)

```bash
go get github.com/nicolasparada/go-db
```

Simple Golang database wrapper over [github.com/jackc/pgx/v5](https://github.com/jackc/pgx) with better transactions API.
Specifically designed to work with [CockroachDB](https://cockroachlabs.com).

Instead of starting a transaction, commiting (or rolling back) each time,
you simply pass a callback function. This allows for defining methods
over a single database object and you can either run them standalone
or inside a transaction.

```go
type Repo struct {
    db *db.DB
}

func (repo *Repo) Insert(ctx context.Context) error {
    repo.db.Exec(ctx, "INSERT INTO ...")
}

func (repo *Repo) Update(ctx context.Context) error {
    repo.db.Exec(ctx, "UPDATE ... SET ...")
}

func (repo *Repo) InsertAndUpdate(ctx context.Context) error {
    return repo.db.RunTx(ctx, func(ctx context.Context) error {
        repo.Insert(ctx)
        repo.Update(ctx)
    })
}
```

## How it works?

When you call `RunTx` it starts a new transaction and saves that object inside context. Calls to `Query`, `QueryRow` and `Exec` will check on the context and will either use the new transaction object or take a connection directly from the pool.
