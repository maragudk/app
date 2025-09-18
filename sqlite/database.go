// Package sqlite provides a [Database], which is a simple wrapper around the stdlib database connection pool at [sql.DB].
// All storage-related functions are methods on [Database].
package sqlite

import (
	"context"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
	"maragu.dev/glue/sql"
)

type Database struct {
	H   *sql.Helper
	log *slog.Logger
}

type NewDatabaseOptions struct {
	H   *sql.Helper
	Log *slog.Logger
}

// NewDatabase with the given options.
// If no logger is provided, logs are discarded.
func NewDatabase(opts NewDatabaseOptions) *Database {
	if opts.Log == nil {
		opts.Log = slog.New(slog.DiscardHandler)
	}

	return &Database{
		H:   opts.H,
		log: opts.Log,
	}
}

func (d *Database) Ping(ctx context.Context) error {
	return d.H.Ping(ctx)
}

type Tx = sql.Tx
