ifneq ("$(wildcard .env)", "")
    include .env
    export $(shell sed 's/=.*//' .env)
endif

.PHONY: install-deps generate test clean

BIN_DIR := $(shell go env GOPATH)/bin
SERVICE_DIR = internal/service
REPO_DIR = internal/repository
UTILS_DIR = internal/utils
PB_DIR = pkg/account_v1

DOCS_DIR = ./docs

tests-build: install-deps generate-pb go-generate-mock

run-tests:
	go test -v ./..

build:  generate compose-up

compose-up:
	docker-compose up -d

install-deps:
	@echo "Installing dependencies..."
	sudo apt-get update && sudo apt-get install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golang/mock/mockgen@latest
	@echo "Dependencies installed successfully!"

generate: swagger generate-pb  go-generate-mock

generate-pb:
	mkdir -p $(PB_DIR)
	protoc --proto_path api/account_v1 \
	--go_out=$(PB_DIR) --go_opt=paths=source_relative \
	--go-grpc_out=$(PB_DIR) --go-grpc_opt=paths=source_relative \
	api/account_v1/account.proto

go-generate-mock:
	$(shell go env GOPATH)/bin/mockgen -destination=$(SERVICE_DIR)/mock/mocks.go -source=$(SERVICE_DIR)/service.go
	$(shell go env GOPATH)/bin/mockgen -destination=$(REPO_DIR)/mock/mocks.go -source=$(REPO_DIR)/repository.go
	$(shell go env GOPATH)/bin/mockgen -destination=$(PB_DIR)/mock/mocks.go -source=$(PB_DIR)/account_grpc.pb.go AccountServiceServer
	$(shell go env GOPATH)/bin/mockgen -destination=$(UTILS_DIR)/mock/mocks.go -source=$(UTILS_DIR)/utils.go

migrate-down:
	docker run --rm -v ./schema:/migrations --network host migrate/migrate \
  -path=/migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" down 1

swagger:
	swag init --output $(DOCS_DIR) --generalInfo ./cmd/http_server/main.go