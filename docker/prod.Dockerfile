FROM golang:1.24.1-alpine AS base

WORKDIR /app

# ==========================
FROM base AS build

COPY --link go.mod go.sum ./

RUN go mod download

COPY --link Makefile ./

RUN apk add make

RUN make docker-install-deps

COPY . .

RUN make generate-docs

RUN go build -o main.exe cmd/http_server/main.go

# ==========================
FROM base

COPY --from=build /app/main.exe /app/main.exe
COPY --from=build /app/docs /app/docs

CMD ["./main.exe"]
