# ==========================
# Загрузка переменных из .env
# ==========================
ifneq ("$(wildcard .env)", "")
	include .env
	export $(shell sed 's/=.*//' .env)
endif

# ==========================
# Общие переменные
# ==========================
BIN_DIR := $(shell go env GOPATH)/bin
SERVICE_DIR = internal/service
REPO_DIR = internal/repository
UTILS_DIR = internal/utils
PB_DIR = pkg/account_v1
DOCS_DIR = ./docs

.PHONY: install-deps generate test clean run-tests build compose-up \
	generate-pb go-generate-mock swagger migrate-up-remote migrate-down tests-build

# ==========================
# Локальная разработка / Тесты
# ==========================

tests-build: install-deps generate-pb go-generate-mock swagger

run-tests:
	@echo "Tidying go.mod and go.sum..."
	go mod tidy
	@echo "Running tests..."
	go test -v ./...

build: generate compose-up

compose-up:
	docker-compose up -d

# ==========================
# Установка зависимостей
# ==========================

install-deps:
	@echo "Installing dependencies..."
	sudo apt-get update && sudo apt-get install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	export PATH=$$PATH:$(go env GOPATH)/bin
	@echo "Dependencies installed successfully!"

# ==========================
# Генерация кода
# ==========================

generate: swagger generate-pb go-generate-mock

generate-pb:
	mkdir -p $(PB_DIR)
	protoc --proto_path api/account_v1 \
	--go_out=$(PB_DIR) --go_opt=paths=source_relative \
	--go-grpc_out=$(PB_DIR) --go-grpc_opt=paths=source_relative \
	api/account_v1/account.proto

go-generate-mock:
	$(BIN_DIR)/mockgen -destination=$(SERVICE_DIR)/mock/mocks.go -source=$(SERVICE_DIR)/service.go
	$(BIN_DIR)/mockgen -destination=$(REPO_DIR)/mock/mocks.go -source=$(REPO_DIR)/repository.go
	$(BIN_DIR)/mockgen -destination=$(PB_DIR)/mock/mocks.go -source=$(PB_DIR)/account_grpc.pb.go AccountServiceServer
	$(BIN_DIR)/mockgen -destination=$(UTILS_DIR)/mock/mocks.go -source=$(UTILS_DIR)/utils.go

swagger:
	swag init --output $(DOCS_DIR) --generalInfo ./cmd/http_server/main.go

# ==========================
# Миграции: удалённое окружение
# Данные для подключения к удалённой БД берутся из .prod.env
# ==========================

migrate-up-remote:
	@echo "Применение миграций к удалённой БД..."
	@if [ ! -f .prod.env ]; then \
		echo "Файл .prod.env не найден!"; \
		exit 1; \
	fi

	export $(shell grep -v '^#' .prod.env | xargs) && \
	migrate -path ./schema \
	-database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" \
	up

migrate-down-remote:
	@echo "Применение миграций к удалённой БД..."
	@if [ ! -f .prod.env ]; then \
		echo "Файл .prod.env не найден!"; \
		exit 1; \
	fi

	export $(shell grep -v '^#' .prod.env | xargs) && \
	migrate -path ./schema \
	-database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable" \
	down

# ==========================
# Миграции: локальная БД (через docker)
# ==========================

migrate-down:
	docker run --rm \
		-v ./schema:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		down 1