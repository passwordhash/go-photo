FROM golang:1.24.1-alpine AS base

WORKDIR /app

# ==========================
FROM base AS build

COPY --link go.mod go.sum ./

# Устанавливаем зависимости
RUN apk add git make protobuf protobuf-dev

COPY --link Makefile ./
RUN make docker-install-deps

# Клонируем proto-файлы
RUN git clone https://github.com/passwordhash/protobuf-files.git api/

COPY . .

RUN make generate

# Собираем бинарник
RUN go build -o main.exe cmd/http_server/main.go

# ==========================
FROM base

# Только исполняемый файл
COPY --from=build /app/main.exe /app/main.exe

CMD ["./main.exe"]
