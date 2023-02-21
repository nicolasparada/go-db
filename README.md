# Golang Database Wrapper

```bash
go get github.com/nicolasparada/go-db
```

Simple Golang database wrapper over [github.com/jackc/pgx/v5](https://github.com/jackc/pgx/v5) with better transactions API.

Instead of starting a transaction, commiting (or rolling back) each time,
you simply pass a callback function. This allows for definig methods
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
        repo.Insert()
        repo.Update()
    })
}
```
