# Include .env file
include ./.env
export $(shell sed 's/=.*//' .env)

PROJECT_NAME := gateway
PKG := ./...
MAIN := ./cmd/api/

MOCKS_DIR := ./mocks
MOCKERY_CONFIG_PATH := ~/.mockery.yml


.PHONY: all build fmt vet test run migrate-create migrate-up migrate-down migrate-drop swag mocks

all: fmt vet test build

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(PROJECT_NAME).exe $(MAIN)
fmt:
	 go fmt $(PKG)

vet:
	 go vet $(PKG)

test:
	go test -v $(PKG)

run:
	go run $(MAIN)

swag:
	@swag init --generalInfo ./cmd/main.go

mocks:
	mockery --config $(MOCKERY_CONFIG_PATH)