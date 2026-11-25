-include .env

APP_NAME ?= app
DATABASE_PATH ?= app.db

# versitygw is the local S3 server, run with go run so it stays out of go.mod.
VERSITYGW_VERSION = v1.11.0
VERSITYGW = go run github.com/versity/versitygw/cmd/versitygw@$(VERSITYGW_VERSION)
VERSITYGW_DATA_DIR = .versitygw
VERSITYGW_PORT ?= 7070
VERSITYGW_TEST_PORT ?= 7072
VERSITYGW_WEBUI_PORT ?= 7071

.PHONY: benchmark
benchmark:
	go test -tags sqlite_fts5,sqlite_math_functions -bench . ./...

.PHONY: build-docker
build-docker:
	docker build --platform linux/arm64 -t $(APP_NAME) .

.PHONY: clean-all
clean-all:
	rm -rf $(VERSITYGW_DATA_DIR)
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

# Starts a throwaway versitygw for the tests in the background, and stops it afterwards.
# The first run compiles versitygw, which takes a minute or so. go run doesn't pass signals
# on to the program it runs, so the cleanup stops go run's child, the server, with pkill -P.
.PHONY: test
test:
	@curl -s -o /dev/null http://localhost:$(VERSITYGW_TEST_PORT); \
	if [ $$? -ne 7 ]; then \
		echo "Port $(VERSITYGW_TEST_PORT) is in use, so the test S3 server can't start there" >&2; \
		exit 1; \
	fi; \
	echo "Starting versitygw on port $(VERSITYGW_TEST_PORT) for the tests"; \
	dir=$$(mktemp -d); \
	$(VERSITYGW) --access access --secret secretsecret --port :$(VERSITYGW_TEST_PORT) --health /health \
		posix "$$dir" >versitygw-test.log 2>&1 & \
	pid=$$!; \
	trap 'pkill -P $$pid || kill $$pid 2>/dev/null; wait $$pid; rm -rf "$$dir"' EXIT; \
	trap 'exit 130' INT TERM; \
	deadline=$$(( $$(date +%s) + 300 )); \
	until curl -fs -o /dev/null http://localhost:$(VERSITYGW_TEST_PORT)/health; do \
		if ! kill -0 $$pid 2>/dev/null || [ $$(date +%s) -ge $$deadline ]; then \
			echo "The test S3 server didn't start, see versitygw-test.log:" >&2; \
			cat versitygw-test.log >&2; \
			exit 1; \
		fi; \
		sleep 0.2; \
	done; \
	go test -tags sqlite_fts5,sqlite_math_functions -coverprofile cover.out -shuffle on ./...

# Runs versitygw for development in the foreground, started by make watch.
.PHONY: versitygw
versitygw:
	mkdir -p $(VERSITYGW_DATA_DIR)
	$(VERSITYGW) --access access --secret secretsecret --port :$(VERSITYGW_PORT) \
		--webui :$(VERSITYGW_WEBUI_PORT) --webui-no-tls posix $(VERSITYGW_DATA_DIR)

.PHONY: watch
watch: tailwindcss
	go tool redo
