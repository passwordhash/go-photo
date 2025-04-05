FROM golang:1.24.1-alpine AS base

WORKDIR /app

# ==========================
FROM base AS build

COPY --link go.mod go.sum ./

# Устанавливаем зависимости
RUN apk add git make protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Клонируем proto-файлы и Makefile
RUN git clone https://github.com/passwordhash/protobuf-files.git api/
COPY --link Makefile ./

# Генерируем protobuf
RUN make generate-pb

COPY . .

# Собираем бинарник
RUN go build -o main.exe cmd/http_server/main.go

# ==========================
FROM base

# Только исполняемый файл
COPY --from=build /app/main.exe /app/main.exe

CMD ["./main.exe"]