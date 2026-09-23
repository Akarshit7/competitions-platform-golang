# Talenzap API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Go backend with DB+Redis, admin email/password auth, phone OTP login, home feeds, and social APIs using only net/http.

**Architecture:** Layered handler/service/store with pgxpool, go-redis, golang-jwt, bcrypt, stdlib mux. Migrations via golang-migrate.

**Tech Stack:** Go 1.26, pgx/v5, pgxpool, go-redis/v9, golang-migrate, golang-jwt/jwt/v5, google/uuid, bcrypt

**Spec:** `docs/superpowers/specs/2026-09-18-talenzap-api-design.md`

## Global Constraints
- Use `net/http` only (no chi/gin/echo) — Go 1.22+ `HandleFunc("GET /path")` patterns
- Go 1.26, module `github.com/Akarshit7/competition-platform-go`
- Env from `.env` via `internal/config`
- Postgres via `DATABASE_URL`, Redis via `REDIS_HOST/PORT/PASSWORD`
- JWT secrets/expiries from env (JWT_TOKEN_SECRET etc.)
- Pagination via DEFAULT_PAGE / DEFAULT_PAGE_SIZE

## Review Focus
- Invalid phone formats still create OTP → expect 400
- Expired OTP in Redis → expect 401 invalid/expired
- Admin login with wrong password → expect 401 not 500
- Vote value not 1/-1 → expect 400
- Missing Authorization header on protected routes → expect 401

---

### Task 1: Extend config + add deps

**Files:**
- Modify: `internal/config/config.go`
- Modify: `go.mod`

**Interfaces:**
- Consumes: os env
- Produces: Config fields `RedisAddr, RedisPassword, DefaultPage, DefaultPageSize, CorsOrigins`

- [ ] Step 1: Add deps `go get github.com/jackc/pgx/v5 github.com/redis/go-redis/v9 github.com/golang-jwt/jwt/v5 github.com/google/uuid golang.org/x/crypto/bcrypt github.com/golang-migrate/migrate/v4`
- [ ] Step 2: Extend Config struct with Redis fields, pagination, cors; update MustLoad()
- [ ] Step 3: Run `go mod tidy && go build ./...` passes

### Task 2: DB + Redis connections

**Files:**
- Create: `internal/database/postgres.go` — NewPool(cfg) (pgxpool.Config parse, ping)
- Create: `internal/redis/redis.go` — NewClient(cfg) (redis.Options, ping)

- [ ] Step 1: Implement postgres.go with context timeout 5s
- [ ] Step 2: Implement redis.go with go-redis, TLS disabled, DB 0
- [ ] Step 3: Verify `go build ./...`

### Task 3: Migrations + schema + seed

**Files:**
- Create: `migrations/000001_init.up.sql`
- Create: `migrations/000001_init.down.sql`
- Create: `cmd/seed/main.go`

- [ ] Step 1: Write up.sql for admins, users, videos, submissions, votes, likes, comments, comment_replies (uuid, timestamps)
- [ ] Step 2: Write down.sql
- [ ] Step 3: Implement seed that inserts admin (email admin@talenzap.com, password Admin@123 hashed) if not exists
- [ ] Step 4: Test `go run cmd/seed/main.go` (needs DB)

### Task 4: Auth tokens + middleware (net/http)

**Files:**
- Create: `internal/auth/jwt.go` (GenerateAccess, GenerateRefresh, Verify)
- Create: `internal/middleware/auth.go` (AuthMiddleware, AdminMiddleware)
- Create: `internal/middleware/cors.go`
- Create: `internal/response/json.go`

- [ ] Step 1: jwt.go with claims {user_id, role, exp}
- [ ] Step 2: middleware extracts Bearer, verifies via auth, puts userID+role in context
- [ ] Step 3: json helper WriteJSON, Error

### Task 5: Admin auth handler/service/store

**Files:**
- Create: `internal/models/admin.go`
- Create: `internal/store/admin_store.go`
- Create: `internal/service/admin_service.go`
- Create: `internal/handler/admin_handler.go`

Endpoints: POST /api/v1/admin/auth/login

### Task 6: User OTP auth

**Files:**
- Create: `internal/models/user.go`
- Create: `internal/store/user_store.go`
- Create: `internal/service/otp_service.go`
- Create: `internal/handler/auth_handler.go`

Endpoints: POST /api/v1/auth/otp/send, POST /api/v1/auth/otp/verify, POST /api/v1/auth/refresh, GET /api/v1/auth/me

### Task 7: Home routes + video store

**Files:**
- Create: `internal/store/video_store.go`
- Create: `internal/service/home_service.go`
- Create: `internal/handler/home_handler.go`

Endpoints: GET /api/v1/home/latest, /trending, /new

### Task 8: Social routes

**Files:**
- Create: `internal/store/submission_store.go` etc. (vote, like, comment)
- Create: `internal/handler/social_handler.go`
- Modify: `cmd/api/main.go` wire all handlers via ServeMux

Endpoints: submissions, vote, like, comment, reply

### Task 9: Wire main.go & verify

**Files:**
- Modify: `cmd/api/main.go`

- [ ] Step: NewServeMux, register all HandleFunc with method prefixes, wrap with CORS, logging, recovery, listen on cfg.Port
- [ ] Verify `go build ./... && go vet ./...`

