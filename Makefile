APP_NAME := app

.PHONY: benchmark
benchmark:
	go test -bench . ./...

.PHONY: build-css
build-css: tailwindcss
		./tailwindcss -i tailwind.css -o public/styles/app.css --minify

.PHONY: build-docker
build-docker:
	docker build --platform linux/arm64 -t $(APP_NAME) .

.PHONY: clean-all
clean-all: down
	docker volume rm $(APP_NAME)_minio

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: down
down:
	@docker compose down minio

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
	go test -coverprofile cover.out -shuffle on ./...

.PHONY: test-down
test-down:
	docker compose down minio-test

.PHONY: test-up
test-up:
	docker compose up -d minio-test

.PHONY: up
up:
	@docker compose up -d minio

.PHONY: watch
watch: up
	./watch.sh

.PHONY: watch-css
watch-css: tailwindcss
	./tailwindcss -i tailwind.css -o public/styles/app.css --watch
