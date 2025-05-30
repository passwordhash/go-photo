# Go Photo
[![Swagger UI](https://img.shields.io/badge/docs-Swagger-blue?logo=swagger)](https://go-photo.passwordhash.tech/api/v1/docs/index.html)
[![codecov](https://codecov.io/gh/passwordhash/go-photo/branch/develop/graph/badge.svg?token=4TW15AUT4C)](https://codecov.io/gh/passwordhash/go-photo)

Микросервис для загрузки, обработки и хранения фотографий, написанный на языке *Go*. Он предоставляет *RESTful API* для загрузки, получения и удаления фотографий. Микросервис использует [внешний SSO-сервис](https://github.com/passwordhash/go-sso) по *gRPC* для выполнения аутентификации и авторизации пользователей.

---

## Зависимости проекта

- [Go](https://golang.org/) версии 1.24 или выше
- Упомянутый выше go-sso, который предоставляет *gRPC* API для аутентификации и авторизации пользователей
- [Библиотека](https://github.com/passwordhash/protos) с моими proto файлами и сгенерированными кодами, моки
- БД: PostgreSQL версии 15 или выше и миграции с Migrate
- Генерация: swagger, mockgen

## Develop развертывание в Docker

> Установка всех зависимостей происходит в отдельном `build stage` в [Dockerfile](Dockerfile), что позволяет избежать установки зависимостей на локальной машине. Также в [Makefile](Makefile) прописаны команды для генерации кода, миграций, сборки проекта, запуска тестов и др.

- Склонировать репозиторий
    ```
    git clone https://github.com/passwordhash/go-photo.git
    cd go-photo
    ```

- Заполнить файл `.env` на основе [.env.example](.env.example)
    ```
    cp docker/.env.example docker/.env
    ```

- Поднять проект
    ```
    make upp
    ```
    или вручную (p.s. команда может быть неактульной)
    ```
    docker-compose -p go-photo -f docker/docker-compose.yml up -d
    ```

## Описание CI/CD

### Непрерывная интеграция (CI)

При разработке этого проекта я научился настраивать и использовать процессы непрерывной интеграции (CI).

При любом push/pull request запускается GitHub Actions, который выполняет следующие шаги:
1. Сборка проекта на удаленном сервере.
2. Запуск тестов.
3. Публикация в Docker Hub.

[//]: # (3. Проверка кода на соответствие стандартам Go.)

### Непрерывная доставка (CD)

Кроме того, я освоил принципы непрерывной доставки (CD), которые позволяют автоматически развертывать приложение после успешного прохождения всех этапов CI. Это обеспечивает более быстрый и надежный процесс доставки новых версий приложения на сервер.

При каждом push/pull request в master запускается GitHub Actions, который выполняет следующие шаги:
1. Подключение к удаленному серверу по SSH.
2. Получение секретов из развернутого `HashiCorp Vault`.
3. Получение Docker образа из Docker Hub и его запуск с переменными окружения.

> Для работы с секретами используеются `Github Secrets` и `HashiCorp Vault`, что обеспечивает безопасность и конфиденциальность данных.
