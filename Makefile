.PHONY: build test vet run-service up down logs tidy fmt

build:
	go build ./...

test:
	go test ./... -race -cover

vet:
	go vet ./...

run-service:
	go run ./cmd/service

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f service

tidy:
	go mod tidy
	go fmt ./...

fmt:
	go fmt ./...
	go mod tidy