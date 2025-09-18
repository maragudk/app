package jobs

import (
	"log/slog"

	"maragu.dev/glue/email/postmark"
	"maragu.dev/glue/jobs"

	"app/model"
)

type RegisterOpts struct {
	Log    *slog.Logger
	Sender *postmark.Sender
}

// Register all available jobs with the given dependencies.
func Register(r *jobs.Runner, opts RegisterOpts) {
	if opts.Log == nil {
		opts.Log = slog.New(slog.DiscardHandler)
	}

	r.Register(model.JobNameSendEmail.String(), SendEmail(opts.Log, opts.Sender))
}
