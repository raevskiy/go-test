# Simple CRUD Interface

This repository is designed for roasting, in educational purpose. The task can be found [here](/TASK.md)

## Prerequisites

- [Docker](https://www.docker.com/get-started/)
- [Goose](https://github.com/pressly/goose)
- [Gosec](https://github.com/securego/gosec)
- Create .env file from .env.example. Default values are ok.

## Just run it locally

```
## Via Docker
docker compose up -d
```

## Setting up the development environment

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
```

3. Run application

```
go run cmd/main.go
```