# Go Photo
[![Swagger UI](https://img.shields.io/badge/docs-Swagger-blue?logo=swagger)](https://go-photo.passwordhash.tech/api/v1/docs/index.html)

## Develop развертывание

- Склонировать репозиторий
    ```
    git clone git@github.com:passwordhash/go-photo.git
    cd go-photo
    ```
  
- Установить зависимости
    ```
    go mod tidy
    ```

- Склонировать репозиторий с proto-файлами
    ```
    git clone https://github.com/passwordhash/protobuf-files.git api/
    ```

- Собрать проект и запустить
    ```
    make build 
    go run cmd/server/main.go 
    ```