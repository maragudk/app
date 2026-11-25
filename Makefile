-include .env

APP_NAME ?= app
DATABASE_PATH ?= app.db

# versitygw is the local S3 server, built into bin/ with go install so it stays out of go.mod.
# The version is pinned here, once, and is part of the binary's path, so changing it rebuilds.
# After changing it, run make test-down, so the next make test starts the new version.
VERSITYGW_VERSION = v1.11.0
VERSITYGW = bin/versitygw-$(VERSITYGW_VERSION)/versitygw
VERSITYGW_PORT ?= 7070
VERSITYGW_WEBUI_PORT ?= 7071

# Each checkout runs its own test server, so tests in parallel worktrees don't share one.
# Its port comes from the checkout's path. Set VERSITYGW_TEST_PORT in .env to pick another.
VERSITYGW_TEST_PORT ?= $(shell pwd | cksum | awk '{ print 20000 + $$1 % 10000 }')

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
	AWS_ENDPOINT_URL=http://localhost:$(VERSITYGW_TEST_PORT) go test -tags sqlite_fts5,sqlite_math_functions -coverprofile cover.out -shuffle on ./...

.PHONY: test-down
test-down:
	@kill $$(cat .versitygw-test/pid 2>/dev/null) 2>/dev/null || true
	rm -rf .versitygw-test

# Starts this checkout's versitygw for the tests in the background, unless it's already running.
# It keeps running after the tests, until make test-down.
.PHONY: test-up
test-up: $(VERSITYGW)
	@if kill -0 $$(cat .versitygw-test/pid 2>/dev/null) 2>/dev/null && curl -fs --max-time 1 -o /dev/null http://localhost:$(VERSITYGW_TEST_PORT)/health; then exit; fi; \
	if curl -s --max-time 1 -o /dev/null http://localhost:$(VERSITYGW_TEST_PORT); then echo "Port $(VERSITYGW_TEST_PORT) is in use, set VERSITYGW_TEST_PORT in .env to a free port" >&2; exit 1; fi; \
	echo "Starting versitygw on port $(VERSITYGW_TEST_PORT) for the tests"; \
	mkdir -p .versitygw-test/data; \
	$(VERSITYGW) --access access --secret secretsecret --port :$(VERSITYGW_TEST_PORT) --health /health posix .versitygw-test/data >versitygw-test.log 2>&1 & echo $$! >.versitygw-test/pid; \
	for i in $$(seq 20); do curl -fs --max-time 1 -o /dev/null http://localhost:$(VERSITYGW_TEST_PORT)/health && exit; sleep 0.5; done; \
	echo "The test S3 server didn't start, see versitygw-test.log:" >&2; cat versitygw-test.log >&2; exit 1

# Runs versitygw for development in the foreground, started by make watch.
.PHONY: versitygw
versitygw: $(VERSITYGW)
	mkdir -p .versitygw
	$(VERSITYGW) --access access --secret secretsecret --port :$(VERSITYGW_PORT) \
		--webui :$(VERSITYGW_WEBUI_PORT) --webui-no-tls posix .versitygw

$(VERSITYGW):
	GOBIN="$(CURDIR)/$(@D)" go install github.com/versity/versitygw/cmd/versitygw@$(VERSITYGW_VERSION)

.PHONY: watch
watch: tailwindcss
	go tool redo
