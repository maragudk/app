// Package servicetest provides testing helpers for the service package.
package servicetest

import (
	"testing"

	"maragu.dev/glue/email/postmarktest"
	"maragu.dev/glue/s3test"

	"app/service"
	"app/sqlitetest"
)

type NewFatOption func(*newFatOptions)

type newFatOptions struct {
	dbOpts []sqlitetest.NewDatabaseOption
}

// WithSQLiteTestOptions passes options to the underlying sqlitetest.NewDatabase call.
func WithSQLiteTestOptions(opts ...sqlitetest.NewDatabaseOption) NewFatOption {
	return func(o *newFatOptions) {
		o.dbOpts = append(o.dbOpts, opts...)
	}
}

// NewFat for testing, with optional options.
func NewFat(t *testing.T, opts ...NewFatOption) *service.Fat {
	t.Helper()

	o := &newFatOptions{}
	for _, opt := range opts {
		opt(o)
	}

	return service.NewFat(service.NewFatOptions{
		Bucket: s3test.NewBucket(t),
		Database: sqlitetest.NewDatabase(t, o.dbOpts...),
		Sender: postmarktest.NewSender(t),
	})
}
