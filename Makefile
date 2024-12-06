include .env

generate: generate-user-api

generate-user-api:
	mkdir -p pkg/account_v1
	protoc --proto_path api/account_v1 \
	--go_out=pkg/account_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/account_v1 --go-grpc_opt=paths=source_relative \
	api/account_v1/account.proto

migrate-down:
	docker run --rm -v ./schema:/migrations --network host migrate/migrate \
  -path=/migrations -database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" down 1
