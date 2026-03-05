// Package servicetest provides testing helpers for the service package.
package servicetest

import (
	"testing"

	"maragu.dev/glue/s3test"

	"app/service"
	"app/sqlitetest"
)

// NewFat for testing, with optional options passed to the underlying database.
func NewFat(t *testing.T, opts ...sqlitetest.NewDatabaseOption) *service.Fat {
	t.Helper()

	return &service.Fat{
		Bucket: s3test.NewBucket(t),
		DB:     sqlitetest.NewDatabase(t, opts...),
	}
}
