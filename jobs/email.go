package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maragu.dev/glue/jobs"

	"app/model"
)

type emailSender interface {
	SendTransactional(ctx context.Context, name string, emailAddress model.EmailAddress, subject, preheader, templateName string, kw model.Keywords) error
}

func SendEmail(log *slog.Logger, sender emailSender) jobs.Func {
	return jobs.WithTracing("jobs.SendEmail", func(ctx context.Context, m []byte) error {
		var jd model.SendEmailJobData
		if err := json.Unmarshal(m, &jd); err != nil {
			panic(err)
		}

		trace.SpanFromContext(ctx).SetAttributes(attribute.String("email.type", jd.Type))

		log.Info("Sending email", "type", jd.Type, "email", jd.Email)

		var err error
		switch jd.Type {
		case "login":
			err = sendLoginEmail(sender, ctx, jd)
		default:
			panic("unknown email type " + jd.Type)
		}
		if err != nil {
			return err
		}
		return nil
	})
}

func sendLoginEmail(sender emailSender, ctx context.Context, jd model.SendEmailJobData) error {
	subject := fmt.Sprintf("Welcome, %v!", jd.Name)
	return sender.SendTransactional(ctx, jd.Name, jd.Email, subject, "Click the link to log in.", "login", jd.Keywords)
}
