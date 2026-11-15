# Chirpy

A small Twitter-like JSON API in Go: users sign up, log in with JWTs, and post short messages called *chirps*. Postgres stores the data; queries are generated with [sqlc](https://sqlc.dev) and migrations run with [goose](https://github.com/pressly/goose).

## Requirements

- Go 1.27+
- PostgreSQL
- goose (to run migrations)

## Setup

Create a `.env` file in the project root:

```env
DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
PORT=8080
TOKEN_SECRET=change-me
POLKA_KEY=your-polka-key
PLATFORM=dev
```

Run the migrations and start the server:

```sh
goose -dir sql/schema postgres "$DB_URL" up
go run .
```

Check that it's up:

```sh
curl http://localhost:8080/api/healthz
```

## API

See [docs/API.md](docs/API.md) for every endpoint.

## Tests

```sh
go test ./...
```
