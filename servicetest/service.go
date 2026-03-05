// Package servicetest provides testing helpers for the service package.
package servicetest

import (
	"testing"

	"maragu.dev/env"
	"maragu.dev/glue/email"
	"maragu.dev/glue/email/postmark"
	"maragu.dev/glue/s3test"

	"app/model"
	"app/service"
	"app/sqlitetest"
)

// NewFat for testing, with optional options passed to the underlying database.
func NewFat(t *testing.T, opts ...sqlitetest.NewDatabaseOption) *service.Fat {
	t.Helper()

	_ = env.Load("../.env.test")

	return &service.Fat{
		Bucket: s3test.NewBucket(t),
		DB:     sqlitetest.NewDatabase(t, opts...),
		Sender: postmark.NewSender(postmark.NewSenderOptions{
			AppName:                   env.GetStringOrDefault("APP_NAME", "App"),
			BaseURL:                   env.GetStringOrDefault("BASE_URL", "http://localhost:8080"),
			Emails:                    email.GetTemplates(),
			Key:                       env.GetStringOrDefault("POSTMARK_KEY", ""),
			MarketingEmailAddress:     model.EmailAddress(env.GetStringOrDefault("MARKETING_EMAIL_ADDRESS", "marketing@example.com")),
			MarketingEmailName:        env.GetStringOrDefault("MARKETING_EMAIL_NAME", "Marketing"),
			ReplyToEmailAddress:       model.EmailAddress(env.GetStringOrDefault("REPLY_TO_EMAIL_ADDRESS", "support@example.com")),
			ReplyToEmailName:          env.GetStringOrDefault("REPLY_TO_EMAIL_NAME", "App"),
			TransactionalEmailAddress: model.EmailAddress(env.GetStringOrDefault("TRANSACTIONAL_EMAIL_ADDRESS", "support@example.com")),
			TransactionalEmailName:    env.GetStringOrDefault("TRANSACTIONAL_EMAIL_NAME", "App"),
		}),
	}
}
