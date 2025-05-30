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

.PHONY: install-deps generate test clean run-tests up compose-up \
	generate-pb generate-mock swagger migrate-up-remote migrate-down tests-build

# ==========================
# Локальная разработка / Тесты
# ==========================

up:
	docker-compose \
		--env-file .env \
		-p go-photo \
		-f docker/docker-compose.yml \
		up -d

tests-build: install-deps generate-mock generate-mock

integration-test:
	@echo "Запуск интеграционных тестов..."
	go test -v ./tests/...

run-tests:
	@echo "Установка go зависимостей..."
	go mod tidy
	@echo "Запуск unit тестов..."
	go test -v `go list ./... | grep -v '/tests'`

EXCLUDE_PATTERNS = model error converter mock handler.go service.go repository.go
# Кроссплатформенная подмена команды sed
SED_INPLACE = $(shell uname | grep -q Darwin && echo "sed -i ''" || echo "sed -i")
collect-coverage-ci:
	@echo "Сбор покрытия для CI..."
	go test -coverprofile=coverage_raw.out -v \
		./internal/handler/v1/... \
		./internal/service/... \
		./internal/repository/...
#		./internal/handler/v1/photos/.. \
#		./internal/handler/v1/public/ \
#		./internal/service/photo \
#		./internal/service/auth \
#		./internal/repository/photo

	@echo "Фильтрация лишних файлов из покрытия..."
	@cp coverage_raw.out coverage.out
	@$(foreach pattern,$(EXCLUDE_PATTERNS), $(SED_INPLACE) "/${pattern}/d" coverage.out;)

	@echo "Покрытие:"
	@go tool cover -func=coverage.out

collect-coverage:
	@echo "Сбор покрытия..."
	go test -coverprofile=coverage_raw.out \
		./internal/handler/v1/auth/ \
		./internal/handler/v1/photos/ \
		./internal/handler/v1/user/ \
		./internal/handler/v1/public/ \
		./internal/service/photo \
		./internal/service/user \
		./internal/repository/photo

	@echo "Результаты покрытия:"
	@go tool cover -func=coverage.out

	@echo "HTML-отчёт:"
	@go tool cover -html=coverage.out

# ==========================
# Установка зависимостей
# ==========================

install-deps:
	@echo "Установка зависимостей..."
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	export PATH=$$PATH:$(go env GOPATH)/bin
	@echo "Установка зависимостей завершена"

# ==========================
# Генерация кода
# ==========================

generate: generate-mock generate-docs

generate-mock:
	$(BIN_DIR)/mockgen -destination=$(SERVICE_DIR)/mock/mocks.go -source=$(SERVICE_DIR)/interface.go
	$(BIN_DIR)/mockgen -destination=$(REPO_DIR)/mock/mocks.go -source=$(REPO_DIR)/interface.go
#	$(BIN_DIR)/mockgen -destination=$(PB_DIR)/mock/mocks.go -source=$(PB_DIR)/account_grpc.pb.go AccountServiceServer
	$(BIN_DIR)/mockgen -destination=$(UTILS_DIR)/mock/mocks.go -source=$(UTILS_DIR)/interface.go

generate-docs:
	swag init --output $(DOCS_DIR) --generalInfo ./cmd/http_server/main.go

# ==========================
# Миграции: удалённое окружение
# Данные для подключения к удалённой БД берутся из .prod.env
# ==========================

migrate-up:
	docker run --rm \
		-v ./migrations:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		up

migrate-down:
	docker run --rm \
		-v ./migrations:/migrations \
		--network host migrate/migrate \
		-path=/migrations \
		-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		down 1

# ===========================
# Dockerfile build
# ===========================

docker-install-deps:
	echo "Установка зависимостей внутрь docker контейнера..."
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	echo "Установка зависимостей завершена"
