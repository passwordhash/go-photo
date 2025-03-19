FROM golang:1.23-alpine AS base

WORKDIR /app

# ==========================
FROM base AS build

COPY --link go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main.exe cmd/http_server/main.go

# ==========================
FROM base

COPY --from=build /app/main.exe /app/main.exe

VOLUME /app/logs

CMD ["./main.exe"]
