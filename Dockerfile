FROM golang:1.24.1-alpine AS base

WORKDIR /app

# ==========================
FROM base AS build

COPY --link go.mod go.sum ./

RUN go mod download

# COPY --link Makefile ./
# RUN make docker-install-deps

COPY . .

# Собираем бинарник
RUN go build -o main.exe cmd/http_server/main.go

# ==========================
FROM base

COPY --from=build /app/main.exe /app/main.exe

CMD ["./main.exe"]
