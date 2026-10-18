# Talenzap Backend

A high-performance REST API for a short-video competition platform, written in pure Go using only the standard library's `net/http` router alongside `pgx` and `go-redis`.

Built around phone-based OTP authentication for end users and email/password authentication for admins, it powers home feeds, competitive submissions, voting, likes, and threaded comments.

---

## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Data Model](#data-model)
- [API Reference](#api-reference)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Usage](#usage)
- [Design Notes](#design-notes)

---

## Overview

Talenzap is the backend for a competition-driven video platform. Users sign in with a phone number and a one-time password, browse curated feeds, submit videos to competitions, vote on submissions, and interact through likes and nested comments.

Key characteristics:

- **Zero web frameworks** — routing uses Go 1.22+ method-and-pattern `net/http.ServeMux`.
- **Layered design** — handlers, stores, and infrastructure are cleanly separated.
- **Graceful degradation** — the server boots even if Postgres or Redis is unavailable, and dependent routes return `503` instead of crashing.
- **Structured logging** — `log/slog` with request-scoped request IDs.

---

## Tech Stack

| Concern | Choice |
| --- | --- |
| Language | Go 1.26 |
| HTTP routing | Standard library `net/http` (Go 1.22 pattern router) |
| Database | PostgreSQL via `pgx/v5` (`pgxpool`) |
| Cache / OTP store | Redis via `go-redis/v9` |
| Auth | JWT (`golang-jwt/jwt/v5`), HS256 |
| Passwords | `bcrypt` (`golang.org/x/crypto`) |
| Config | `godotenv` + environment variables |
| Logging | `log/slog` (text in dev, JSON in prod) |

---

## Architecture

The application follows a layered, dependency-injected architecture. `main` wires together configuration, infrastructure clients, stores, and handlers, then registers routes on a single `ServeMux`.

```
                          +---------------------+
          HTTP request ->  |   Global middleware |
                          |  RequestLogger      |
                          |    -> Recovery      |
                          |      -> CORS        |
                          +----------+----------+
                                     |
                          +----------v----------+
                          |  ServeMux (routes)  |
                          +----------+----------+
                                     |
              +----------------------+----------------------+
              |                      |                      |
       +------v------+       +-------v------+       +-------v------+
       | Auth        |       | Auth         |       | Admin        |
       | Middleware  |       | handlers     |       | handlers     |
       | (JWT check) |       | (OTP, me)    |       | (login, me)  |
       +------+------+       +-------+------+       +-------+------+
              |                      |                      |
              |              +-------v------+               |
              |              | Home / Social|               |
              |              | handlers     |               |
              |              +-------+------+               |
              |                      |                      |
              +----------+-----------+----------------------+
                         |
                 +-------v--------+
                 |     Stores     |
                 | Admin, User,   |
                 | Video, Social  |
                 +---+--------+---+
                     |        |
             +-------v--+  +--v--------+
             | Postgres |  |  Redis    |
             | (pgxpool)|  | (go-redis)|
             +----------+  +-----------+
```

**Layers**

1. **Middleware** — request logging (assigns/propagates a request ID), panic recovery, and CORS. Auth middleware validates JWTs and injects the user ID into the request context.
2. **Handlers** — transport-layer logic: decode/validate input, call stores, and emit JSON. No SQL lives here.
3. **Stores** — one per aggregate (`AdminStore`, `UserStore`, `VideoStore`, `SocialStore`), encapsulating all SQL and returning domain models.
4. **Infrastructure** — `database` (pgxpool) and `redis` (client) packages manage connections and health.

---

## Project Structure

```
.
├── cmd/
│   ├── api/main.go          # Application entrypoint: wiring + routes
│   └── seed/main.go         # Seeds the default admin account
├── internal/
│   ├── auth/                # JWT generation and verification
│   ├── config/              # Env loading and validation
│   ├── database/            # Postgres connection pool
│   ├── handler/             # HTTP handlers (admin, auth, home, social)
│   ├── logger/              # slog setup + request ID context helpers
│   ├── middleware/          # CORS, auth, request logger, recovery
│   ├── models/              # Domain structs (User, Admin, Video, ...)
│   ├── redis/               # Redis client construction
│   ├── response/            # JSON response helpers
│   └── store/               # Data-access layer
├── migrations/              # Up/down SQL migrations
├── Makefile                 # build / run / clean targets
└── go.mod
```

---

## Data Model

The schema (see `migrations/000001_init.up.sql`) centers on users, videos, and competitions:

- **admins** — email, bcrypt `password_hash`, name.
- **users** — phone (unique), optional name.
- **videos** — title, description, url, thumbnail, owner, `view_count`.
- **submissions** — a user's entry into a competition, optionally linked to a video, with a `status`.
- **votes** — `+1` / `-1` per user per submission (unique constraint).
- **likes** — one per user per video (unique constraint).
- **comments** and **comment_replies** — threaded discussion on videos.

Indexes cover hot paths: recent videos, view counts, and per-entity lookups.

---

## API Reference

All routes are prefixed with `/api/v1`. Protected routes require an `Authorization: Bearer <access_token>` header.

### Health

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/health` | Public | Liveness check |
| GET | `/api/v1/health` | Public | Liveness check |

### Admin

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/admin/auth/login` | Public | Email/password login, returns JWT pair |
| GET | `/admin/me` | Admin | Current admin profile |

### Auth (users)

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/auth/otp/send` | Public | Send a 6-digit OTP to a phone |
| POST | `/auth/otp/verify` | Public | Verify OTP, upsert user, issue tokens |
| POST | `/auth/refresh` | Public | Exchange a refresh token for new tokens |
| GET | `/auth/me` | User | Current user profile |

### Home feeds

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/home/latest` | Public | Latest videos |
| GET | `/home/trending` | Public | Most-viewed videos |
| GET | `/home/new` | Public | Newest videos |

Feed endpoints accept `?page=` and `?limit=` (defaults from config, max 100).

### Social

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/submissions` | User | Create a competition submission |
| GET | `/submissions` | Public | List submissions |
| POST | `/submissions/{id}/vote` | User | Vote `+1` or `-1` |
| POST | `/videos/{id}/like` | User | Toggle a like |
| POST | `/videos/{id}/comments` | User | Create a comment |
| GET | `/videos/{id}/comments` | Public | List comments |
| POST | `/comments/{id}/replies` | User | Reply to a comment |
| GET | `/comments/{id}/replies` | Public | List replies |

---

## Getting Started

### Prerequisites

- Go 1.26+
- PostgreSQL 14+
- Redis 6+

### 1. Clone

```bash
git clone git@github.com:Akarshit7/competitions-platform-golang.git
cd competitions-platform-golang
```

### 2. Configure

Create a `.env` file in the project root (see [Environment Variables](#environment-variables)).

### 3. Run migrations

Apply the SQL in `migrations/` to your database, e.g.:

```bash
psql "$DATABASE_URL" -f migrations/000001_init.up.sql
```

### 4. Seed the admin

```bash
go run ./cmd/seed
```

### 5. Run the API

```bash
make run-api
# or
go run ./cmd/api
```

The server listens on `PORT` (default from env) and exposes `GET /health`.

---

## Environment Variables

| Variable | Required | Description |
| --- | --- | --- |
| `PORT` | Yes | HTTP listen port |
| `ENV` / `NODE_ENV` | Yes | `development` (text logs) or production (JSON logs) |
| `DATABASE_URL` | Yes | Postgres DSN |
| `REDIS_HOST` | No | Defaults to `127.0.0.1` |
| `REDIS_PORT` | No | Defaults to `6379` |
| `REDIS_PASSWORD` | No | Redis auth (optional) |
| `JWT_TOKEN_SECRET` | Yes | Access-token signing secret |
| `REFRESH_TOKEN_SECRET` | Yes | Refresh-token signing secret |
| `TEMP_TOKEN_SECRET` | Yes | Temporary-token signing secret |
| `JWT_TOKEN_EXPIRY` | Yes | Access-token TTL (seconds) |
| `REFRESH_TOKEN_EXPIRY` | Yes | Refresh-token TTL (seconds) |
| `TEMP_TOKEN_EXPIRY` | Yes | Temporary-token TTL (seconds) |
| `ADMIN_JWT_TOKEN_EXPIRY` | Yes | Admin access-token TTL (seconds) |
| `DEFAULT_PAGE` | No | Default page number (fallback `1`) |
| `DEFAULT_PAGE_SIZE` | No | Default page size (fallback `10`) |
| `CORS_ORIGINS` | No | Comma-separated allowed origins |

Config is validated at startup; the process panics on missing required values.

---

## Usage

### Build

```bash
make build-api      # outputs bin/api
make clean          # removes the binary
```

### Example: OTP login flow

```bash
curl -X POST localhost:8080/api/v1/auth/otp/send \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+919999999999"}'

# In development the OTP is written to the structured logs.
curl -X POST localhost:8080/api/v1/auth/otp/verify \
  -H 'Content-Type: application/json' \
  -d '{"phone":"+919999999999","otp":"123456"}'
```

### Example: authenticated request

```bash
curl localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

## Design Notes

- **Graceful degradation.** `main` logs warnings when Postgres or Redis is unreachable but still starts. Stores are `nil` in that case, and handlers return `503 Service Unavailable` for dependent routes.
- **OTP delivery is mocked.** In development the OTP is emitted via structured logs; swap in an SMS provider for production.
- **JWT strategy.** Access and refresh tokens are signed with separate secrets. Admin tokens use their own expiry so admin sessions can be tuned independently.
- **Pagination.** `page`/`limit` are parsed defensively, capped at 100, with configurable defaults.
- **Request IDs.** The logger middleware attaches a request ID to context so log lines across a request can be correlated.
