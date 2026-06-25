-include .env

APP_NAME ?= app
AWS_ACCESS_KEY_ID ?= access
AWS_ENDPOINT_URL ?= http://localhost:7070
AWS_REGION ?= us-east-1
AWS_SECRET_ACCESS_KEY ?= secretsecret
DATABASE_PATH ?= app.db
S3_BUCKET_NAME ?= bucket

.PHONY: benchmark
benchmark:
	go test -tags sqlite_fts5,sqlite_math_functions -bench . ./...

.PHONY: bucket
bucket:
	@set -e; \
	cfg=$$(mktemp); \
	trap 'rm -f $$cfg' EXIT; \
	printf '[default]\ns3 =\n    addressing_style = path\n' >$$cfg; \
	export AWS_CONFIG_FILE=$$cfg \
		AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		AWS_DEFAULT_REGION=$(AWS_REGION); \
	echo "Waiting for versitygw at $(AWS_ENDPOINT_URL)..."; \
	for i in $$(seq 1 30); do \
		if aws --endpoint-url $(AWS_ENDPOINT_URL) s3api list-buckets >/dev/null 2>&1; then break; fi; \
		if [ $$i -eq 30 ]; then echo "versitygw not reachable at $(AWS_ENDPOINT_URL) after 30s" >&2; exit 1; fi; \
		sleep 1; \
	done; \
	if aws --endpoint-url $(AWS_ENDPOINT_URL) s3api head-bucket --bucket $(S3_BUCKET_NAME) >/dev/null 2>&1; then \
		echo "Bucket $(S3_BUCKET_NAME) already exists"; \
	else \
		aws --endpoint-url $(AWS_ENDPOINT_URL) s3api create-bucket --bucket $(S3_BUCKET_NAME) >/dev/null; \
		echo "Created bucket $(S3_BUCKET_NAME)"; \
	fi

.PHONY: build-docker
build-docker:
	docker build --platform linux/arm64 -t $(APP_NAME) .

.PHONY: clean-all
clean-all: down
	docker volume rm $(APP_NAME)_versitygw
	rm -f $(DATABASE_PATH) $(DATABASE_PATH)-wal $(DATABASE_PATH)-shm

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: deps
deps:
	curl -Lf -o public/scripts/datastar.js https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.8/bundles/datastar.js

.PHONY: down
down:
	docker compose down versitygw

.PHONY: fmt
fmt:
	goimports -w -local `head -n 1 go.mod | sed 's/^module //'` .

.PHONY: lint
lint:
	golangci-lint run

tailwindcss:
	curl -sfL -o tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod a+x tailwindcss

.PHONY: test
test: test-up
	go test -tags sqlite_fts5,sqlite_math_functions -coverprofile cover.out -shuffle on ./...

.PHONY: test-down
test-down:
	docker compose down versitygw-test

.PHONY: test-up
test-up:
	docker compose up -d versitygw-test

.PHONY: up
up:
	docker compose up -d versitygw
	$(MAKE) bucket

.PHONY: watch
watch: tailwindcss
	go tool redo
