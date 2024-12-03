include .env

generate:
	make generate-user-api

generate-user-api:
	mkdir -p pkg/account_v1
	protoc --proto_path api/account_v1 \
	--go_out=pkg/account_v1 --go_opt=paths=source_relative \
	--go-grpc_out=pkg/account_v1 --go-grpc_opt=paths=source_relative \
	api/account_v1/account.proto
