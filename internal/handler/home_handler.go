package handler

import (
	"net/http"
	"strconv"

	"github.com/Akarshit7/competition-platform-go/internal/config"
	"github.com/Akarshit7/competition-platform-go/internal/response"
	"github.com/Akarshit7/competition-platform-go/internal/store"
)

type HomeHandler struct {
	cfg        config.Config
	videoStore *store.VideoStore
}

func NewHomeHandler(cfg config.Config, vs *store.VideoStore) *HomeHandler {
	return &HomeHandler{cfg: cfg, videoStore: vs}
}

func (h *HomeHandler) parsePagination(r *http.Request) (limit, offset int) {
	limit = h.cfg.DefaultPageSize
	offset = 0
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

func (h *HomeHandler) Latest(w http.ResponseWriter, r *http.Request) {
	if h.videoStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	limit, offset := h.parsePagination(r)
	videos, err := h.videoStore.ListLatest(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch videos")
		return
	}
	if videos == nil {
		response.JSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}, "limit": limit, "offset": offset})
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": videos, "limit": limit, "offset": offset})
}

func (h *HomeHandler) Trending(w http.ResponseWriter, r *http.Request) {
	if h.videoStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	limit, offset := h.parsePagination(r)
	videos, err := h.videoStore.ListTrending(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch trending")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": videos, "limit": limit, "offset": offset})
}

func (h *HomeHandler) NewVideos(w http.ResponseWriter, r *http.Request) {
	if h.videoStore == nil {
		response.Error(w, http.StatusServiceUnavailable, "db not connected")
		return
	}
	limit, offset := h.parsePagination(r)
	videos, err := h.videoStore.ListNew(r.Context(), limit, offset)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch new videos")
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": videos, "limit": limit, "offset": offset})
}
