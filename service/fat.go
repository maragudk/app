package service

import (
	"context"

	"maragu.dev/glue/email/postmark"

	"app/model"
	"app/sqlite"
)

type Fat struct {
	DB     *sqlite.Database
	Sender *postmark.Sender
	// add more as needed
}

func (f *Fat) Signup(ctx context.Context, input model.SignupInput) error {
	// placeholder: business logic goes here
	return nil
}
