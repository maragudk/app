package jobs_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"maragu.dev/errors"
	"maragu.dev/glue/email"
	"maragu.dev/is"

	"app/jobs"
	"app/model"
)

type senderStub struct {
	err  error
	opts email.SendOptions
}

func (s *senderStub) SendTransactional(ctx context.Context, opts email.SendOptions) error {
	s.opts = opts
	return s.err
}

func TestSendEmail(t *testing.T) {
	t.Run("should send a login email with the recipient, subject, preheader, template, and keywords", func(t *testing.T) {
		sender := &senderStub{}

		is.NotError(t, jobs.SendEmail(newLog(), sender)(t.Context(), marshalJobData(t, model.SendEmailJobData{
			Type:     "login",
			Name:     "Me",
			Email:    "me@example.com",
			Keywords: model.Keywords{"token": "123"},
		})))

		is.Equal(t, "Me", sender.opts.ToName)
		is.Equal(t, model.EmailAddress("me@example.com"), sender.opts.To)
		is.Equal(t, "Welcome, Me!", sender.opts.Subject)
		is.Equal(t, "Click the link to log in.", sender.opts.Preheader)
		is.Equal(t, "login", sender.opts.Template)
		is.Equal(t, "123", sender.opts.Keywords["token"])
	})

	t.Run("should leave the reply-to unset, so the sender uses the one it is configured with", func(t *testing.T) {
		sender := &senderStub{}

		is.NotError(t, jobs.SendEmail(newLog(), sender)(t.Context(), marshalJobData(t, model.SendEmailJobData{
			Type:  "login",
			Name:  "Me",
			Email: "me@example.com",
		})))

		is.Equal(t, model.EmailAddress(""), sender.opts.ReplyTo)
		is.Equal(t, "", sender.opts.ReplyToName)
	})

	t.Run("should return the error if the sender fails", func(t *testing.T) {
		sendErr := errors.New("no")
		sender := &senderStub{err: sendErr}

		err := jobs.SendEmail(newLog(), sender)(t.Context(), marshalJobData(t, model.SendEmailJobData{
			Type:  "login",
			Name:  "Me",
			Email: "me@example.com",
		}))
		is.Error(t, sendErr, err)
	})

	t.Run("should panic on an unknown email type", func(t *testing.T) {
		defer func() {
			is.Equal(t, "unknown email type nope", recover())
		}()

		_ = jobs.SendEmail(newLog(), &senderStub{})(t.Context(), marshalJobData(t, model.SendEmailJobData{
			Type:  "nope",
			Email: "me@example.com",
		}))

		t.Fatal("did not panic")
	})
}

func newLog() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func marshalJobData(t *testing.T, jd model.SendEmailJobData) []byte {
	t.Helper()

	m, err := json.Marshal(jd)
	is.NotError(t, err)
	return m
}
