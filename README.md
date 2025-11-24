# Simple CRUD Interface

Rewrite the README according to the application.

The task itself can be found [here](/TASK.md)

## Prerequisites

- [Docker](https://www.docker.com/get-started/)
- [Goose](https://github.com/pressly/goose)
- [Gosec](https://github.com/securego/gosec)

## Getting Started

0. Create .env file from .env.example

1. Start database

```
## Via Makefile
make db

## Via Docker
docker compose -f docker-compose.dev.yml up -d
```

2. Run migrations

```
## Via Makefile
make migrate-up

## Via Goose
DB_DRIVER=postgres
DB_STRING="host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
goose -dir ./migrations $(DB_DRIVER) $(DB_STRING) up
```

3. Run application

```
go run cmd/main.go
```