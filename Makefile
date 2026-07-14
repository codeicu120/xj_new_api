APP_NAME := xj-comp-api
GO_PACKAGES := ./...

.PHONY: test
test:
	go test $(GO_PACKAGES)

.PHONY: lint
lint:
	go vet $(GO_PACKAGES)

.PHONY: run
run:
	go run ./cmd/api

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP_NAME) ./cmd/api

.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME):local .

.PHONY: ci
ci: test lint build
