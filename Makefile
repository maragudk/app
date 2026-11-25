-include .env

APP_NAME ?= app
DATABASE_PATH ?= app.db

# versitygw is the local S3 server, built into bin/ with go install so it stays out of go.mod.
# The version is pinned here, once, and is part of the binary's path, so changing it rebuilds.
VERSITYGW_VERSION = v1.11.0
VERSITYGW = bin/versitygw-$(VERSITYGW_VERSION)/versitygw
VERSITYGW_PORT ?= 7070
VERSITYGW_WEBUI_PORT ?= 7071

# Each checkout runs its own test server, so tests in parallel worktrees don't share one. Its id
# comes from the checkout's path and is part of its health path, which is how test-up and
# test-down tell it apart from other servers. The id also picks where its port search starts.
VERSITYGW_TEST_ID = $(shell pwd | cksum | cut -d ' ' -f 1)
VERSITYGW_TEST_HEALTH = /health-$(VERSITYGW_TEST_ID)

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
	AWS_ENDPOINT_URL=http://localhost:$$(cat .versitygw-test/port) go test -tags sqlite_fts5,sqlite_math_functions -coverprofile cover.out -shuffle on ./...

.PHONY: test-down
test-down:
	@pkill -f -F .versitygw-test/pid '$(VERSITYGW_TEST_HEALTH) ' 2>/dev/null; \
	for i in $$(seq 50); do pgrep -f -F .versitygw-test/pid '$(VERSITYGW_TEST_HEALTH) ' >/dev/null 2>&1 || break; sleep 0.1; done; \
	pkill -KILL -f -F .versitygw-test/pid '$(VERSITYGW_TEST_HEALTH) ' 2>/dev/null || true
	rm -rf .versitygw-test

# Starts this checkout's versitygw for the tests in the background, unless it's already running.
# It keeps running after the tests, until make test-down, and restarts after a version change.
# The port search skips ports where something listens (curl exits with 7 when nothing does).
.PHONY: test-up
test-up: $(VERSITYGW)
	@if pgrep -f -F .versitygw-test/pid '^$(VERSITYGW) .*$(VERSITYGW_TEST_HEALTH) ' >/dev/null 2>&1 && \
		curl -fs --max-time 2 -o /dev/null http://localhost:$$(cat .versitygw-test/port)$(VERSITYGW_TEST_HEALTH); then exit; fi; \
	pkill -f -F .versitygw-test/pid '$(VERSITYGW_TEST_HEALTH) ' 2>/dev/null; \
	port=$$((20000 + $(VERSITYGW_TEST_ID) % 10000)); \
	while curl -s --max-time 1 -o /dev/null http://localhost:$$port; [ $$? -ne 7 ]; do port=$$((port + 1)); done; \
	echo "Starting versitygw on port $$port for the tests"; \
	mkdir -p .versitygw-test/data; \
	echo $$port >.versitygw-test/port; \
	$(VERSITYGW) --access access --secret secretsecret --port :$$port --health $(VERSITYGW_TEST_HEALTH) \
		posix .versitygw-test/data >versitygw-test.log 2>&1 & echo $$! >.versitygw-test/pid; \
	for i in $$(seq 150); do \
		curl -fs --max-time 1 -o /dev/null http://localhost:$$port$(VERSITYGW_TEST_HEALTH) && exit; \
		kill -0 $$! 2>/dev/null || break; \
		sleep 0.2; \
	done; \
	kill $$! 2>/dev/null; \
	echo "The test S3 server didn't start, see versitygw-test.log:" >&2; \
	cat versitygw-test.log >&2; \
	exit 1

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
