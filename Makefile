include .env

SERVICE_MOCKDIR = internal/service/mocks
SERVICE_MOCKGEN_SRC = internal/service/service.go
REPO_MOCKDIR = internal/repository/mocks
REPO_MOCKGEN_SRC = internal/repository/repository.go

DOCS_DIR = ./docs

build:  generate compose-up

compose-up:
	docker-compose up -d

generate: swagger generate-pb  go-generate-mock

generate-pb:
	mkdir -p pkg/account_v1
	protoc --proto_path api/account_v1 \
	--go_out=pkg/account_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/account_v1 --go-grpc_opt=paths=source_relative \
	api/account_v1/account.proto

go-generate-mock:
	$(GOPATH)/bin/mockgen -destination=$(SERVICE_MOCKDIR)/mock.go -source=$(SERVICE_MOCKGEN_SRC)
	$(GOPATH)/bin/mockgen -destination=$(REPO_MOCKDIR)/mock.go -source=$(REPO_MOCKGEN_SRC)
	$(GOPATH)/bin/mockgen -destination=pkg/account_v1/mocks/mock.go -source=pkg/account_v1/account_grpc.pb.go AccountServiceServer


migrate-down:
	docker run --rm -v ./schema:/migrations --network host migrate/migrate \
  -path=/migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" down 1

swagger:
	swag init --output $(DOCS_DIR) --generalInfo ./cmd/http_server/main.go