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
  
## Описание CI/CD 

### Непрерывная интеграция (CI)

При разработке проекта go-photo я научился настраивать и использовать процессы непрерывной интеграции (CI). 

При любом push/pull request запускается GitHub Actions, который выполняет следующие шаги:
1. Сборка проекта на удаленном сервере.
2. Запуск тестов.

[//]: # (3. Проверка кода на соответствие стандартам Go.)

### Непрерывная доставка (CD)

Кроме того, я освоил принципы непрерывной доставки (CD), которые позволяют автоматически развертывать приложение после успешного прохождения всех этапов CI. Это обеспечивает более быстрый и надежный процесс доставки новых версий приложения на сервер.

    Настройка автоматического развертывания:
        Я настроил рабочие процессы для автоматического развертывания приложения на сервер после успешного прохождения всех тестов.
        Пример конфигурационного файла для автоматического развертывания:
        YAML

name: CD

on:
  push:
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.16

      - name: Build application
        run: go build -o go-photo ./cmd/go-photo

      - name: Deploy to server
        run: |
          scp go-photo user@server:/path/to/deploy
          ssh user@server 'systemctl restart go-photo'

Public code references from 2 repositories

Эти навыки позволили мне автоматизировать весь процесс разработки, начиная от написания кода и заканчивая его развертыванием, что значительно повысило эффективность и качество работы над проектом.

Если у вас есть еще какие-либо аспекты проекта, которые вы хотели бы описать, пожалуйста, сообщите мне!