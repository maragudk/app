package service

import (
	"context"

	"maragu.dev/glue/email/postmark"
	"maragu.dev/glue/s3"

	"app/model"
	"app/sqlite"
)

type Fat struct {
	Bucket *s3.Bucket
	DB     *sqlite.Database
	Sender *postmark.Sender
}

func (f *Fat) Signup(ctx context.Context, input model.SignupInput) error {
	// placeholder: business logic goes here
	return nil
}
