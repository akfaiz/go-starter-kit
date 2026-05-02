# Go API Starter Kit

API-only starter kit built with Go.

## Stack

- Echo v5
- OpenAPI router: `github.com/oaswrap/spec/adapter/echov5openapi`
- Bun + PostgreSQL
- Migris migrations
- JWT auth (access + refresh)
- Forgot password with OTP via email
- go-mailgen for email content
- Uber FX for dependency injection

## Quick Start

```bash
cp .env.example .env
go mod tidy
go run . migrate up
go run . serve
```

Server: `http://localhost:8080`
OpenAPI docs: `http://localhost:8080/docs`

## Auth Endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh-token`
- `POST /api/v1/auth/forgot-password/send-otp`
- `POST /api/v1/auth/forgot-password/verify-otp`
- `POST /api/v1/auth/forgot-password/reset-password`

## Migrations

```bash
go run . migrate status
go run . migrate up
go run . migrate down
```

## Docker

```bash
docker compose up --build
```
