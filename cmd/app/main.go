package main

import (
	"context"
	"log/slog"
	"time"

	"maragu.dev/env"
	"maragu.dev/errors"
	"maragu.dev/glue/app"
	"maragu.dev/glue/aws"
	"maragu.dev/glue/email"
	"maragu.dev/glue/email/postmark"
	gluehttp "maragu.dev/glue/http"
	gluejobs "maragu.dev/glue/jobs"
	"maragu.dev/glue/s3"
	"maragu.dev/glue/sql"
	"maragu.dev/glue/sqlitestore"

	"app/html"
	"app/http"
	"app/jobs"
	"app/model"
	"app/sqlite"
)

func main() {
	app.Start(start)
}

func start(ctx context.Context, log *slog.Logger, eg app.Goer) error {
	databaseLog := log.With("component", "sql.Database")

	jobTimeout := env.GetDurationOrDefault("JOB_QUEUE_TIMEOUT", 10*time.Second)

	db := sqlite.NewDatabase(sqlite.NewDatabaseOptions{
		H: sql.NewHelper(sql.NewHelperOptions{
			JobQueue: sql.JobQueueOptions{
				Timeout: jobTimeout,
			},
			Log: databaseLog,
			SQLite: sql.SQLiteOptions{
				Path: env.GetStringOrDefault("DATABASE_PATH", "app.db"),
			},
		}),
		Log: databaseLog,
	})
	if err := db.H.Connect(ctx); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	if err := db.H.MigrateUp(ctx); err != nil {
		return errors.Wrap(err, "error migrating database")
	}

	awsConfig, err := aws.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	bucket := s3.NewBucket(s3.NewBucketOptions{
		Config:    awsConfig,
		Name:      env.GetStringOrDefault("S3_BUCKET_NAME", "bucket"),
		PathStyle: env.GetBoolOrDefault("S3_PATH_STYLE", false),
	})

	runner := gluejobs.NewRunner(gluejobs.NewRunnerOpts{
		Log:   log.With("component", "jobs.Runner"),
		Queue: db.H.JobsQ,
	})

	baseURL := env.GetStringOrDefault("BASE_URL", "http://localhost:8080")

	sender := postmark.NewSender(postmark.NewSenderOptions{
		AppName:                   env.GetStringOrDefault("APP_NAME", "App"),
		BaseURL:                   baseURL,
		Emails:                    email.GetTemplates(),
		Key:                       env.GetStringOrDefault("POSTMARK_KEY", ""),
		Log:                       log.With("component", "email.Sender"),
		MarketingEmailAddress:     model.EmailAddress(env.GetStringOrDefault("REPLY_TO_EMAIL_ADDRESS", "marketing@example.com")),
		MarketingEmailName:        env.GetStringOrDefault("MARKETING_EMAIL_NAME", "Marketing"),
		ReplyToEmailAddress:       model.EmailAddress(env.GetStringOrDefault("REPLY_TO_EMAIL_ADDRESS", "support@example.com")),
		ReplyToEmailName:          env.GetStringOrDefault("REPLY_TO_EMAIL_NAME", "Support"),
		TransactionalEmailAddress: model.EmailAddress(env.GetStringOrDefault("TRANSACTIONAL_EMAIL_ADDRESS", "transactional@example.com")),
		TransactionalEmailName:    env.GetStringOrDefault("TRANSACTIONAL_EMAIL_NAME", "Transactional"),
	})

	jobs.Register(runner, jobs.RegisterOpts{
		Log:    log.With("component", "jobs"),
		Sender: sender,
	})

	store, err := sqlitestore.New(ctx, db.H.DB.DB)
	if err != nil {
		return errors.Wrap(err, "error creating sqlite session store")
	}

	server := gluehttp.NewServer(gluehttp.NewServerOptions{
		Address:            env.GetStringOrDefault("SERVER_ADDRESS", ":8080"),
		BaseURL:            baseURL,
		CSP:                http.CSP(env.GetBoolOrDefault("CSP_ALLOW_UNSAFE_INLINE", false)),
		HTMLPage:           html.Page,
		HTTPRouterInjector: http.InjectHTTPRouter(log, db, bucket),
		Log:                log.With("component", "http.Server"),
		PermissionsGetter:  db,
		SecureCookie:       env.GetBoolOrDefault("SECURE_COOKIE", true),
		SessionStore:       store,
		UserActiveChecker:  db,
	})

	eg.Go(func() error {
		return server.Start(ctx)
	})

	eg.Go(func() error {
		runner.Start(ctx)
		return nil
	})

	return nil
}
