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
	generate-pb generate-mock swagger migrate-up-remote migrate-down tests-build

# ==========================
# Локальная разработка / Тесты
# ==========================

build: generate
	docker-compose up -d

tests-build: install-deps generate-pb generate-mock generate-mock

run-tests:
	@echo "Установка go зависимостей..."
	go mod tidy
	@echo "Запуск тестов..."
	go test -v ./...

# ==========================
# Установка зависимостей
# ==========================

install-deps:
	@echo "Установка зависимостей..."
	sudo apt-get update && sudo apt-get install -y protobuf-compiler
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	export PATH=$$PATH:$(go env GOPATH)/bin
	@echo "Установка зависимостей завершена"

# ==========================
# Генерация кода
# ==========================

generate: generate-pb generate-mock generate-docs

generate-pb:
	mkdir -p $(PB_DIR)
	protoc --proto_path=api/account_v1 \
		--go_out=$(PB_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PB_DIR) --go-grpc_opt=paths=source_relative \
		api/account_v1/account.proto

generate-mock:
	$(BIN_DIR)/mockgen -destination=$(SERVICE_DIR)/mock/mocks.go -source=$(SERVICE_DIR)/service.go
	$(BIN_DIR)/mockgen -destination=$(REPO_DIR)/mock/mocks.go -source=$(REPO_DIR)/repository.go
	$(BIN_DIR)/mockgen -destination=$(PB_DIR)/mock/mocks.go -source=$(PB_DIR)/account_grpc.pb.go AccountServiceServer
	$(BIN_DIR)/mockgen -destination=$(UTILS_DIR)/mock/mocks.go -source=$(UTILS_DIR)/utils.go

generate-docs:
	swag init --output $(DOCS_DIR) --generalInfo ./cmd/http_server/main.go

# ==========================
# Миграции: удалённое окружение
# Данные для подключения к удалённой БД берутся из .prod.env
# ==========================

migrate-up:
	@echo "Применение миграций к локальной БД..."
	docker run --rm \
		-v ./schema:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		up

migrate-down:
	@echo "Откат миграций к локальной БД..."
	docker run --rm \
		-v ./schema:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		down 1

# ===========================
# Dockerfile build
# ===========================

docker-install-deps:
	echo "Установка зависимостей внутрь docker контейнера..."
	apk add git make protobuf protobuf-dev
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	echo "Установка зависимостей завершена"
