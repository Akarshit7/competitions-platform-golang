package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Akarshit7/competition-platform-go/internal/config"
	"github.com/Akarshit7/competition-platform-go/internal/database"
	"github.com/Akarshit7/competition-platform-go/internal/handler"
	"github.com/Akarshit7/competition-platform-go/internal/logger"
	"github.com/Akarshit7/competition-platform-go/internal/middleware"
	"github.com/Akarshit7/competition-platform-go/internal/response"
	"github.com/Akarshit7/competition-platform-go/internal/store"
	iredis "github.com/Akarshit7/competition-platform-go/internal/redis"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.Env)
	slog.SetDefault(log)

	pool, err := database.NewPool(cfg)
	if err != nil {
		slog.Warn("db connect failed, server will start but DB routes will 500", "error", err)
	} else {
		defer pool.Close()
		slog.Info("db connected", "host", cfg.DatabaseUrl)
		if err := runMigrations(pool); err != nil {
			slog.Warn("migration failed", "error", err)
		}
	}

	rdb, err := iredis.NewClient(cfg)
	if err != nil {
		slog.Warn("redis connect failed, otp will fail", "error", err)
	} else {
		defer rdb.Close()
		slog.Info("redis connected", "addr", cfg.RedisAddr)
	}

	// stores (may be nil if pool nil -> handlers will handle)
	var adminStore *store.AdminStore
	var userStore *store.UserStore
	var videoStore *store.VideoStore
	var socialStore *store.SocialStore
	if pool != nil {
		adminStore = store.NewAdminStore(pool)
		userStore = store.NewUserStore(pool)
		videoStore = store.NewVideoStore(pool)
		socialStore = store.NewSocialStore(pool)
	}

	// handlers
	adminH := handler.NewAdminHandler(cfg, adminStore)
	authH := handler.NewAuthHandler(cfg, userStore, rdb)
	homeH := handler.NewHomeHandler(cfg, videoStore)
	socialH := handler.NewSocialHandler(cfg, socialStore, videoStore)

	mux := http.NewServeMux()

	// health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// admin (public login, protected me)
	mux.HandleFunc("POST /api/v1/admin/auth/login", adminH.Login)
	mux.HandleFunc("GET /api/v1/admin/me", middleware.AdminMiddleware(cfg, http.HandlerFunc(adminH.Me)).ServeHTTP)

	// auth otp
	mux.HandleFunc("POST /api/v1/auth/otp/send", authH.SendOTP)
	mux.HandleFunc("POST /api/v1/auth/otp/verify", authH.VerifyOTP)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)
	mux.HandleFunc("GET /api/v1/auth/me", middleware.AuthMiddleware(cfg, http.HandlerFunc(authH.Me)).ServeHTTP)

	// home
	mux.HandleFunc("GET /api/v1/home/latest", homeH.Latest)
	mux.HandleFunc("GET /api/v1/home/trending", homeH.Trending)
	mux.HandleFunc("GET /api/v1/home/new", homeH.NewVideos)

	// social - submissions
	mux.HandleFunc("POST /api/v1/submissions", middleware.AuthMiddleware(cfg, http.HandlerFunc(socialH.CreateSubmission)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/submissions", socialH.ListSubmissions)
	mux.HandleFunc("POST /api/v1/submissions/{id}/vote", middleware.AuthMiddleware(cfg, http.HandlerFunc(socialH.Vote)).ServeHTTP)

	// likes, comments on videos
	mux.HandleFunc("POST /api/v1/videos/{id}/like", middleware.AuthMiddleware(cfg, http.HandlerFunc(socialH.ToggleLike)).ServeHTTP)
	mux.HandleFunc("POST /api/v1/videos/{id}/comments", middleware.AuthMiddleware(cfg, http.HandlerFunc(socialH.CreateComment)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/videos/{id}/comments", socialH.ListComments)

	// comment replies
	mux.HandleFunc("POST /api/v1/comments/{id}/replies", middleware.AuthMiddleware(cfg, http.HandlerFunc(socialH.CreateReply)).ServeHTTP)
	mux.HandleFunc("GET /api/v1/comments/{id}/replies", socialH.ListReplies)

	// wrap with global middleware: request logging outermost, then recovery, then CORS
	handler := middleware.RequestLogger(middleware.RecoveryWithLog(middleware.CORS(cfg, mux)))

	addr := ":" + cfg.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("server listening", "addr", addr, "env", cfg.Env)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		panic(err)
	}
}

func runMigrations(pool interface{}) error {
	return nil
}
