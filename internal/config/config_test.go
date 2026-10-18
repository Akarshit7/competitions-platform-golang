package config

import (
	"reflect"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("NODE_ENV", "")
	t.Setenv("PORT", "8080")
	t.Setenv("ENV", "test")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/competitions-social-media-platform")
	t.Setenv("JWT_TOKEN_SECRET", "jwt-secret")
	t.Setenv("REFRESH_TOKEN_SECRET", "refresh-secret")
	t.Setenv("TEMP_TOKEN_SECRET", "temp-secret")
	t.Setenv("JWT_TOKEN_EXPIRY", "3600")
	t.Setenv("REFRESH_TOKEN_EXPIRY", "86400")
	t.Setenv("TEMP_TOKEN_EXPIRY", "300")
	t.Setenv("ADMIN_JWT_TOKEN_EXPIRY", "7200")
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic, got none")
		}
	}()
	fn()
}

// 1. Valid env loads and optional values fall back to defaults.
func TestMustLoad_ValidWithDefaults(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("REDIS_HOST", "")
	t.Setenv("REDIS_PORT", "")
	t.Setenv("DEFAULT_PAGE", "")
	t.Setenv("DEFAULT_PAGE_SIZE", "")
	t.Setenv("CORS_ORIGINS", "")

	cfg := MustLoad()

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.Env != "test" {
		t.Errorf("Env = %q, want test", cfg.Env)
	}
	if cfg.RedisHost != "127.0.0.1" || cfg.RedisPort != "6379" {
		t.Errorf("redis = %s:%s, want 127.0.0.1:6379", cfg.RedisHost, cfg.RedisPort)
	}
	if cfg.RedisAddr != "127.0.0.1:6379" {
		t.Errorf("RedisAddr = %q, want 127.0.0.1:6379", cfg.RedisAddr)
	}
	if cfg.DefaultPage != 1 || cfg.DefaultPageSize != 10 {
		t.Errorf("pagination = %d/%d, want 1/10", cfg.DefaultPage, cfg.DefaultPageSize)
	}
	if cfg.JwtTokenExpiry != 3600 || cfg.RefreshTokenExpiry != 86400 || cfg.AdminJwtTokenExpiry != 7200 {
		t.Errorf("expiries = %d/%d/%d, want 3600/86400/7200", cfg.JwtTokenExpiry, cfg.RefreshTokenExpiry, cfg.AdminJwtTokenExpiry)
	}
}

// 2. Missing required variables panic.
func TestMustLoad_MissingRequiredPanics(t *testing.T) {
	required := []string{
		"PORT", "ENV", "DATABASE_URL",
		"JWT_TOKEN_SECRET", "REFRESH_TOKEN_SECRET", "TEMP_TOKEN_SECRET",
	}
	for _, key := range required {
		t.Run(key, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv(key, "")
			assertPanics(t, func() { MustLoad() })
		})
	}
}

// 3. Non-numeric or negative expiry values panic.
func TestMustLoad_InvalidExpiryPanics(t *testing.T) {
	for _, v := range []string{"", "abc", "-1", "3.5"} {
		t.Run(v, func(t *testing.T) {
			setRequiredEnv(t)
			t.Setenv("JWT_TOKEN_EXPIRY", v)
			assertPanics(t, func() { MustLoad() })
		})
	}
}

// 4. CORS origins are split on commas, trimmed, and empty entries dropped.
func TestMustLoad_CORSOrigins(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("CORS_ORIGINS", " http://a.com , http://b.com ,, http://c.com ")

	cfg := MustLoad()

	want := []string{"http://a.com", "http://b.com", "http://c.com"}
	if !reflect.DeepEqual(cfg.CorsOrigins, want) {
		t.Errorf("CorsOrigins = %#v, want %#v", cfg.CorsOrigins, want)
	}
}

// 5. Numeric pagination overrides apply; invalid values fall back to defaults.
func TestMustLoad_PaginationParsing(t *testing.T) {
	t.Run("overrides", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("DEFAULT_PAGE", "3")
		t.Setenv("DEFAULT_PAGE_SIZE", "25")

		cfg := MustLoad()

		if cfg.DefaultPage != 3 || cfg.DefaultPageSize != 25 {
			t.Errorf("pagination = %d/%d, want 3/25", cfg.DefaultPage, cfg.DefaultPageSize)
		}
	})

	t.Run("invalid falls back", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("DEFAULT_PAGE", "not-a-number")
		t.Setenv("DEFAULT_PAGE_SIZE", "abc")

		cfg := MustLoad()

		if cfg.DefaultPage != 1 || cfg.DefaultPageSize != 10 {
			t.Errorf("pagination = %d/%d, want fallback 1/10", cfg.DefaultPage, cfg.DefaultPageSize)
		}
	})
}
