package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Akarshit7/competition-platform-go/internal/auth"
	"github.com/Akarshit7/competition-platform-go/internal/config"
	"github.com/Akarshit7/competition-platform-go/internal/middleware"
	"github.com/Akarshit7/competition-platform-go/internal/response"
	"github.com/Akarshit7/competition-platform-go/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	cfg   config.Config
	store *store.AdminStore
}

func NewAdminHandler(cfg config.Config, s *store.AdminStore) *AdminHandler {
	return &AdminHandler{cfg: cfg, store: s}
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		response.Error(w, http.StatusInternalServerError, "db not connected")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "email and password required")
		return
	}
	admin, err := h.store.FindByEmail(r.Context(), req.Email)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	access, err := auth.GenerateAccessToken(admin.ID, "admin", h.cfg.JwtTokenSecret, h.cfg.AdminJwtTokenExpiry)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	refresh, err := auth.GenerateRefreshToken(admin.ID, "admin", h.cfg.RefreshTokenSecret, h.cfg.RefreshTokenExpiry)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate refresh token")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  access,
		"refresh_token": refresh,
		"admin": map[string]string{
			"id":    admin.ID,
			"email": admin.Email,
			"name":  admin.Name,
		},
	})
}

func (h *AdminHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		// fallback parse token
		hdr := r.Header.Get("Authorization")
		if strings.HasPrefix(hdr, "Bearer ") {
			if claims, err := auth.VerifyToken(strings.TrimPrefix(hdr, "Bearer "), h.cfg.JwtTokenSecret); err == nil {
				userID = claims.UserID
			}
		}
	}
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	admin, err := h.store.FindByID(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "admin not found")
		return
	}
	response.JSON(w, http.StatusOK, admin)
}
