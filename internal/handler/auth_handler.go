package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Akarshit7/competition-platform-go/internal/auth"
	"github.com/Akarshit7/competition-platform-go/internal/config"
	"github.com/Akarshit7/competition-platform-go/internal/middleware"
	"github.com/Akarshit7/competition-platform-go/internal/response"
	"github.com/Akarshit7/competition-platform-go/internal/store"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

type AuthHandler struct {
	cfg       config.Config
	userStore *store.UserStore
	rdb       *redis.Client
}

func NewAuthHandler(cfg config.Config, us *store.UserStore, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{cfg: cfg, userStore: us, rdb: rdb}
}

func generateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	val := n.Int64() + 100000
	return fmt.Sprintf("%06d", val)
}

func (h *AuthHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if !phoneRegex.MatchString(req.Phone) {
		response.Error(w, http.StatusBadRequest, "invalid phone format")
		return
	}
	if h.rdb == nil {
		response.Error(w, http.StatusServiceUnavailable, "redis not connected")
		return
	}
	if h.userStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	otp := generateOTP()
	key := "otp:" + req.Phone
	ctx := r.Context()
	if err := h.rdb.Set(ctx, key, otp, 5*time.Minute).Err(); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to store otp")
		return
	}
	// Mock SMS: log to console (in real, call SMS provider)
	log.Printf("[OTP] phone=%s otp=%s", req.Phone, otp)
	response.JSON(w, http.StatusOK, map[string]string{"message": "otp sent"})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if h.rdb == nil || h.userStore == nil {
		response.Error(w, http.StatusInternalServerError, "service not ready")
		return
	}
	var req struct {
		Phone string `json:"phone"`
		OTP   string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if !phoneRegex.MatchString(req.Phone) {
		response.Error(w, http.StatusBadRequest, "invalid phone format")
		return
	}
	if req.OTP == "" {
		response.Error(w, http.StatusBadRequest, "otp required")
		return
	}
	key := "otp:" + req.Phone
	ctx := r.Context()
	stored, err := h.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		response.Error(w, http.StatusUnauthorized, "otp expired or not found")
		return
	} else if err != nil {
		response.Error(w, http.StatusInternalServerError, "redis error")
		return
	}
	if stored != req.OTP {
		response.Error(w, http.StatusUnauthorized, "invalid otp")
		return
	}
	// delete after verify
	h.rdb.Del(ctx, key)

	user, err := h.userStore.UpsertByPhone(ctx, req.Phone)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	access, err := auth.GenerateAccessToken(user.ID, "user", h.cfg.JwtTokenSecret, h.cfg.JwtTokenExpiry)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	refresh, err := auth.GenerateRefreshToken(user.ID, "user", h.cfg.RefreshTokenSecret, h.cfg.RefreshTokenExpiry)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  access,
		"refresh_token": refresh,
		"user":          user,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		response.Error(w, http.StatusBadRequest, "refresh_token required")
		return
	}
	claims, err := auth.VerifyToken(req.RefreshToken, h.cfg.RefreshTokenSecret)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}
	access, err := auth.GenerateAccessToken(claims.UserID, claims.Role, h.cfg.JwtTokenSecret, h.cfg.JwtTokenExpiry)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	newRefresh, err := auth.GenerateRefreshToken(claims.UserID, claims.Role, h.cfg.RefreshTokenSecret, h.cfg.RefreshTokenExpiry)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"access_token":  access,
		"refresh_token": newRefresh,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.userStore.FindByID(context.Background(), userID)
	if err != nil {
		// try with request context
		user, err = h.userStore.FindByID(r.Context(), userID)
		if err != nil {
			response.Error(w, http.StatusNotFound, "user not found")
			return
		}
	}
	response.JSON(w, http.StatusOK, user)
}
