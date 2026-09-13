package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string
	DatabaseUrl string
	// RedisHost string
	// RedisPort string
	// RedisPassword string
	// RedisTlsEnabled bool
	// DefaultPage uint64
	// DefaultPageSize uint64
	JwtTokenSecret      string
	RefreshTokenSecret  string
	TempTokenSecret     string
	JwtTokenExpiry      uint64
	RefreshTokenExpiry  uint64
	TempTokenExpiry     uint64
	AdminJwtTokenExpiry uint64
}

func MustLoad() Config {
	godotenv.Load()

	port := os.Getenv("PORT")
	env := os.Getenv("ENV")
	if strings.TrimSpace(env) == "" {
		env = os.Getenv("NODE_ENV")
	}
	databaseUrl := os.Getenv("DATABASE_URL")
	jwtTokenSecret := os.Getenv("JWT_TOKEN_SECRET")
	refreshTokenSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	tempTokenSecret := os.Getenv("TEMP_TOKEN_SECRET")
	jwtTokenExpiry := os.Getenv("JWT_TOKEN_EXPIRY")
	refreshTokenExpiry := os.Getenv("REFRESH_TOKEN_EXPIRY")
	tempTokenExpiry := os.Getenv("TEMP_TOKEN_EXPIRY")
	adminJwtTokenExpiry := os.Getenv("ADMIN_JWT_TOKEN_EXPIRY")

	if strings.TrimSpace(port) == "" {
		panic("PORT should not be empty")
	}
	if strings.TrimSpace(env) == "" {
		panic("ENV should not be empty")
	}
	if strings.TrimSpace(databaseUrl) == "" {
		panic("DATABASE_URL should not be empty")
	}
	if strings.TrimSpace(jwtTokenSecret) == "" {
		panic("JWT_TOKEN_SECRET should not be empty")
	}
	if strings.TrimSpace(refreshTokenSecret) == "" {
		panic("REFRESH_TOKEN_SECRET should not be empty")
	}
	if strings.TrimSpace(tempTokenSecret) == "" {
		panic("TEMP_TOKEN_SECRET should not be empty")
	}

	return Config{
		Port:                port,
		Env:                 env,
		DatabaseUrl:         databaseUrl,
		JwtTokenSecret:      jwtTokenSecret,
		RefreshTokenSecret:  refreshTokenSecret,
		TempTokenSecret:     tempTokenSecret,
		JwtTokenExpiry:      mustParseUint64("JWT_TOKEN_EXPIRY", jwtTokenExpiry),
		RefreshTokenExpiry:  mustParseUint64("REFRESH_TOKEN_EXPIRY", refreshTokenExpiry),
		TempTokenExpiry:     mustParseUint64("TEMP_TOKEN_EXPIRY", tempTokenExpiry),
		AdminJwtTokenExpiry: mustParseUint64("ADMIN_JWT_TOKEN_EXPIRY", adminJwtTokenExpiry),
	}
}

func mustParseUint64(name, value string) uint64 {
	if strings.TrimSpace(value) == "" {
		panic(name + " should not be empty")
	}

	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("%s must be a valid unsigned integer: %v", name, err))
	}

	return parsed
}
