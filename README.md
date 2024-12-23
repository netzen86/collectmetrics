# project-template

Шаблон репозитория для проекта по курсу «Продвинутый Go-разработчик для сетевых инженеров».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/mipt-golang-course/project-template.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/mipt-golang-course/go-autotests).


* Формат файла конфигурации для сервера:
```
{
    "address": "localhost:8080", // аналог переменной окружения ADDRESS или флага -a
    "restore": true, // аналог переменной окружения RESTORE или флага -r
    "store_interval": "1s", // аналог переменной окружения STORE_INTERVAL или флага -i
    "store_file": "/path/to/file.db", // аналог переменной окружения STORE_FILE или -f
    "database_dsn": "", // аналог переменной окружения DATABASE_DSN или флага -d
    "crypto_key": "/path/to/key.pem" // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
}
```
* Формат файла конфигурации для агента:

```
{
    "address": "localhost:8080", // аналог переменной окружения ADDRESS или флага -a
    "report_interval": "1s", // аналог переменной окружения REPORT_INTERVAL или флага -r
    "poll_interval": "1s", // аналог переменной окружения POLL_INTERVAL или флага -p
    "crypto_key": "/path/to/key.pem" // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
}
```

* Генерируем go файлы для сервера из topo файла

из корня проекта запускаем данную команду

```
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/server/server.proto
```