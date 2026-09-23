# Talenzap Backend Go - API Design Spec

**Date:** 2026-09-18
**Stack:** Go 1.26, net/http, pgx/v5, golang-migrate, go-redis/v9, JWT (golang-jwt/jwt/v5), bcrypt

## 1. Intent & Scope
Build Go replacement for Node backend with:
- DB (Postgres) + Redis connections
- /admin auth via email+password
- User phone OTP login with JWT + Refresh tokens
- Home routes: latest, trending, new videos
- Social: submission, vote, like, comment, reply
- Constraint: net/http stdlib only (no chi/gin/echo)

Success: server boots, connects DB+Redis, migrations run, admin can login, user can OTP-login, home feeds work, social actions work, all via net/http.

## 2. Architecture
Layered: handler → service → store (pgx) + redis.
- `internal/config` loads env
- `internal/database` pgxpool
- `internal/redis` go-redis client
- `internal/middleware` auth (JWT), cors, logger, recovery
- `internal/auth` token generation/verification
- `internal/models` structs
- `internal/store/*` per entity (admin, user, video, submission, vote, like, comment)
- `internal/handler/*` net/http handlers using `http.ServeMux` with method prefixes (`GET /api/v1/...`)
- `internal/service/*` business logic
- `migrations/` SQL via golang-migrate
- `cmd/api/main.go` wires everything

Routing: `http.NewServeMux()` + `HandleFunc("GET /path", h)` (Go 1.22+ enhanced patterns). No external router.

## 3. Data Model (SQL)
- `admins` id uuid pk, email unique, password_hash, name, created_at
- `users` id uuid pk, phone unique, name nullable, created_at, updated_at
- `videos` id uuid pk, title, description, url, thumbnail_url, user_id fk, view_count int, created_at
- `submissions` id uuid pk, user_id fk, video_id fk, title, description, status, created_at
- `votes` id uuid pk, user_id fk, submission_id fk, value int (1/-1), unique(user,submission)
- `likes` id uuid pk, user_id fk, video_id fk, unique(user,video), created_at
- `comments` id uuid pk, user_id fk, video_id fk, body, created_at
- `comment_replies` id uuid pk, user_id fk, comment_id fk, body, created_at
- `refresh_tokens` id uuid pk, user_id fk, token_hash, expires_at, created_at (or stateless JWT refresh)

Trending = order by (view_count + likes + votes) desc recent 7d. Latest = created_at desc. New = created_at desc limit with pagination.

## 4. Auth Details
- Admin: POST /api/v1/admin/auth/login {email,password} → bcrypt compare → JWT (AdminJwtTokenExpiry, secret JWT_TOKEN_SECRET) + include role=admin. Middleware checks Authorization: Bearer.
- User OTP: POST /api/v1/auth/otp/send {phone} → generate 6-digit, store Redis key `otp:<phone>` 5min, log. POST /api/v1/auth/otp/verify {phone,otp} → check Redis, delete, upsert user, issue accessToken (JWT_TOKEN_EXPIRY) + refreshToken (REFRESH_TOKEN_EXPIRY) signed with respective secrets. POST /api/v1/auth/refresh {refresh_token} → verify → new tokens. GET /api/v1/auth/me requires auth.

## 5. Home Routes
- GET /api/v1/home/latest?page=&limit=
- GET /api/v1/home/trending?page=&limit=
- GET /api/v1/home/new?page=&limit=  (new = latest variant, or filtered last 48h)
All public, paginated via DEFAULT_PAGE/SIZE, return videos with user info.

## 6. Social Routes (auth required except read)
- POST /api/v1/submissions (auth) {video_id,title,description}
- GET /api/v1/submissions?page=&limit=
- POST /api/v1/submissions/{id}/vote {value:1|-1}
- POST /api/v1/videos/{id}/like (toggle)
- POST /api/v1/videos/{id}/comments {body}
- GET /api/v1/videos/{id}/comments?page=&limit=
- POST /api/v1/comments/{id}/replies {body}
- GET /api/v1/comments/{id}/replies
Each mutating checks auth user from context.

## 7. Cross-cutting
- Errors: JSON {error: message}, status codes 400/401/403/404/500
- Pagination helper, validation, logging via slog
- Seed admin via `go run cmd/seed` or migration seed

## 8. Verification
- `go build ./...` passes
- `curl` admin login, otp send/verify, home feeds, social flows
- Migration up/down works

