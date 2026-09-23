package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Akarshit7/competition-platform-go/internal/config"
	"github.com/Akarshit7/competition-platform-go/internal/middleware"
	"github.com/Akarshit7/competition-platform-go/internal/response"
	"github.com/Akarshit7/competition-platform-go/internal/store"
)

type SocialHandler struct {
	cfg         config.Config
	socialStore *store.SocialStore
	videoStore  *store.VideoStore
}

func NewSocialHandler(cfg config.Config, ss *store.SocialStore, vs *store.VideoStore) *SocialHandler {
	return &SocialHandler{cfg: cfg, socialStore: ss, videoStore: vs}
}

func (h *SocialHandler) parsePagination(r *http.Request) (limit, offset int) {
	limit = h.cfg.DefaultPageSize
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	page := h.cfg.DefaultPage
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	offset = (page - 1) * limit
	return limit, offset
}

// POST /api/v1/submissions
func (h *SocialHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil || h.videoStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		VideoID     *string `json:"video_id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		response.Error(w, http.StatusBadRequest, "title required")
		return
	}
	if req.VideoID != nil && *req.VideoID != "" {
		ok, err := h.videoStore.Exists(r.Context(), *req.VideoID)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, "db error")
			return
		}
		if !ok {
			response.Error(w, http.StatusBadRequest, "video not found")
			return
		}
	} else {
		req.VideoID = nil
	}
	sub, err := h.socialStore.CreateSubmission(r.Context(), userID, req.VideoID, req.Title, req.Description)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create submission")
		return
	}
	response.JSON(w, http.StatusCreated, sub)
}

// GET /api/v1/submissions
func (h *SocialHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	limit, offset := h.parsePagination(r)
	subs, err := h.socialStore.ListSubmissions(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list submissions")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": subs, "limit": limit, "offset": offset})
}

// POST /api/v1/submissions/{id}/vote
func (h *SocialHandler) Vote(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		// fallback for Go <1.22 or manual parse
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			id = parts[3]
		}
	}
	if id == "" {
		response.Error(w, http.StatusBadRequest, "submission id required")
		return
	}
	ok, err := h.socialStore.SubmissionExists(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "db error")
		return
	}
	if !ok {
		response.Error(w, http.StatusNotFound, "submission not found")
		return
	}
	var req struct {
		Value int `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Value != 1 && req.Value != -1 {
		response.Error(w, http.StatusBadRequest, "value must be 1 or -1")
		return
	}
	if err := h.socialStore.Vote(r.Context(), userID, id, req.Value); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to vote")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "voted"})
}

// POST /api/v1/videos/{id}/like
func (h *SocialHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil || h.videoStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	videoID := r.PathValue("id")
	if videoID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		// /api/v1/videos/{id}/like => parts[3] is id
		if len(parts) >= 4 {
			videoID = parts[3]
		}
	}
	if videoID == "" {
		response.Error(w, http.StatusBadRequest, "video id required")
		return
	}
	ok, err := h.videoStore.Exists(r.Context(), videoID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "db error")
		return
	}
	if !ok {
		response.Error(w, http.StatusNotFound, "video not found")
		return
	}
	liked, err := h.socialStore.ToggleLike(r.Context(), userID, videoID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to toggle like")
		return
	}
	response.JSON(w, http.StatusOK, map[string]bool{"liked": liked})
}

// POST /api/v1/videos/{id}/comments
func (h *SocialHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil || h.videoStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	videoID := r.PathValue("id")
	if videoID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 4 {
			videoID = parts[3]
		}
	}
	if videoID == "" {
		response.Error(w, http.StatusBadRequest, "video id required")
		return
	}
	ok, err := h.videoStore.Exists(r.Context(), videoID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "db error")
		return
	}
	if !ok {
		response.Error(w, http.StatusNotFound, "video not found")
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		response.Error(w, http.StatusBadRequest, "body required")
		return
	}
	comment, err := h.socialStore.CreateComment(r.Context(), userID, videoID, req.Body)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create comment")
		return
	}
	response.JSON(w, http.StatusCreated, comment)
}

// GET /api/v1/videos/{id}/comments
func (h *SocialHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	videoID := r.PathValue("id")
	if videoID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 4 {
			videoID = parts[3]
		}
	}
	limit, offset := h.parsePagination(r)
	comments, err := h.socialStore.ListComments(r.Context(), videoID, limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list comments")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": comments, "limit": limit, "offset": offset})
}

// POST /api/v1/comments/{id}/replies
func (h *SocialHandler) CreateReply(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	userID := middleware.GetUserID(r)
	if userID == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	commentID := r.PathValue("id")
	if commentID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 4 {
			commentID = parts[3]
		}
	}
	if commentID == "" {
		response.Error(w, http.StatusBadRequest, "comment id required")
		return
	}
	ok, err := h.socialStore.CommentExists(r.Context(), commentID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "db error")
		return
	}
	if !ok {
		response.Error(w, http.StatusNotFound, "comment not found")
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		response.Error(w, http.StatusBadRequest, "body required")
		return
	}
	reply, err := h.socialStore.CreateReply(r.Context(), userID, commentID, req.Body)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create reply")
		return
	}
	response.JSON(w, http.StatusCreated, reply)
}

// GET /api/v1/comments/{id}/replies
func (h *SocialHandler) ListReplies(w http.ResponseWriter, r *http.Request) {
	if h.socialStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	commentID := r.PathValue("id")
	if commentID == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 4 {
			commentID = parts[3]
		}
	}
	replies, err := h.socialStore.ListReplies(r.Context(), commentID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list replies")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": replies})
}
