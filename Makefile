-include .env

APP_NAME ?= app
DATABASE_PATH ?= app.db

.PHONY: benchmark
benchmark:
	go test -tags sqlite_fts5,sqlite_math_functions -bench . ./...

.PHONY: build-docker
build-docker:
	docker build --platform linux/arm64 -t $(APP_NAME) .

.PHONY: clean-all
clean-all: test-down
	rm -rf .versitygw
	rm -f $(DATABASE_PATH) $(DATABASE_PATH)-wal $(DATABASE_PATH)-shm

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: deps
deps:
	curl -Lf -o public/scripts/datastar.js https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.3/bundles/datastar.js

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
	AWS_ENDPOINT_URL=http://localhost:$$(./versitygw.sh port versitygw-test) go test -tags sqlite_fts5,sqlite_math_functions -coverprofile cover.out -shuffle on ./...

.PHONY: test-down
test-down:
	./versitygw.sh down

# Starts this checkout's versitygw for the tests in the background, unless it's already running.
# Each checkout gets its own port, so tests in parallel worktrees don't share one.
# It keeps running after the tests, until make test-down.
.PHONY: test-up
test-up:
	@./versitygw.sh up

.PHONY: watch
watch: tailwindcss
	go tool redo
