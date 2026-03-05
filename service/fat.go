// Package service provides business logic.
// See https://www.alexedwards.net/blog/the-fat-service-pattern
package service

import (
	"context"

	"maragu.dev/glue/email/postmark"
	"maragu.dev/glue/s3"

	"app/model"
	"app/sqlite"
)

type Fat struct {
	bucket *s3.Bucket
	db     *sqlite.Database
	sender *postmark.Sender
}

type NewFatOptions struct {
	Bucket *s3.Bucket
	Database *sqlite.Database
	Sender *postmark.Sender
}

func NewFat(opts NewFatOptions) *Fat {
	return &Fat{
		bucket: opts.Bucket,
		db:     opts.Database,
		sender: opts.Sender,
	}
}

func (f *Fat) GetUser(ctx context.Context, id model.UserID) (model.User, error) {
	return f.db.GetUser(ctx, id)
}
