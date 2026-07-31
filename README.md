# Todo Server

Веб-планировщик задач на Go с хранением данных в SQLite. Приложение позволяет создавать, просматривать, редактировать, выполнять и удалять задачи, а также настраивать их повторение.

Фронтенд расположен в каталоге `web`, серверная часть — в `backend`.

## Выполненные задания повышенной сложности

* расширенные правила повторения `w` и `m`;
* поиск задач по заголовку, комментарию и дате;
* JWT-аутентификация по паролю из `TODO_PASSWORD`;
* сборка и запуск приложения через Docker.

## Локальный запуск

Поддерживаемые переменные окружения:

| Переменная      |   По умолчанию | Назначение                                           |
| --------------- | -------------: | ---------------------------------------------------- |
| `TODO_PORT`     |         `7540` | порт HTTP-сервера                                    |
| `TODO_DBFILE`   | `scheduler.db` | путь к SQLite-базе                                   |
| `TODO_PASSWORD` |          пусто | пароль; при пустом значении аутентификация отключена |

Пример `.env`:

```env
TODO_PORT=7540
TODO_DBFILE=scheduler.db
TODO_PASSWORD=12345
```

Запуск из корня проекта:

```bash
go run ./backend/cmd
```

Либо с явными параметрами:

```bash
TODO_PORT=7540 \
TODO_DBFILE=scheduler.db \
TODO_PASSWORD=12345 \
go run ./backend/cmd
```

Сервис будет доступен по адресу:

```text
http://localhost:7540
```

Логи записываются в `logs/server.log`.

## Запуск через Docker

Docker-образ собирает приложение из исходного кода. Зависимости Go берутся из каталога `vendor`.

Перед первой сборкой подготовьте зависимости:

```bash
go mod tidy
go mod vendor
```

Соберите Docker-образ из корня проекта:

```bash
docker build -t todo-server:latest .
```

Создайте каталоги для базы данных и логов:

```bash
mkdir -p docker-data logs
```

Запустите контейнер:

```bash
docker run --rm \
  --name todo-server \
  -p 7540:7540 \
  -e TODO_PORT=7540 \
  -e TODO_DBFILE=/data/scheduler.db \
  --mount type=bind,src="$(pwd)/docker-data",dst=/data \
  --mount type=bind,src="$(pwd)/logs",dst=/app/logs \
  todo-server:latest
```

Сервис будет доступен по адресу:

```text
http://localhost:7540
```

SQLite-база данных будет храниться на хосте в файле:

```text
docker-data/scheduler.db
```

Логи приложения будут доступны в каталоге:

```text
logs/
```

Для запуска с аутентификацией передайте переменную `TODO_PASSWORD`:

```bash
docker run --rm \
  --name todo-server \
  -p 7540:7540 \
  -e TODO_PORT=7540 \
  -e TODO_DBFILE=/data/scheduler.db \
  -e TODO_PASSWORD=12345 \
  --mount type=bind,src="$(pwd)/docker-data",dst=/data \
  --mount type=bind,src="$(pwd)/logs",dst=/app/logs \
  todo-server:latest
```

Для запуска контейнера в фоновом режиме используйте параметр `-d`:

```bash
docker run -d \
  --name todo-server \
  -p 7540:7540 \
  -e TODO_PORT=7540 \
  -e TODO_DBFILE=/data/scheduler.db \
  --mount type=bind,src="$(pwd)/docker-data",dst=/data \
  --mount type=bind,src="$(pwd)/logs",dst=/app/logs \
  todo-server:latest
```

Посмотреть запущенный контейнер:

```bash
docker ps
```

Остановить контейнер:

```bash
docker stop todo-server
```

