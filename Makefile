APP=calendar
PKG=./...

.PHONY: build run test vet lint tidy

build:
	go build -race -o bin/$(APP) ./cmd/calendar

run:
	PORT=8080 go run ./cmd/calendar

vet:
	go vet $(PKG)

test:
	go test -race -count=1 $(PKG)

lint:
	@echo "Running golangci-lint (if installed)"
	@golangci-lint run || true

tidy:
	go mod tidy
