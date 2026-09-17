package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vency7744/pmos-panel/internal/api"
	"github.com/vency7744/pmos-panel/internal/auth"
	"github.com/vency7744/pmos-panel/internal/config"
	"github.com/vency7744/pmos-panel/internal/filemanager"
	"github.com/vency7744/pmos-panel/internal/journald"
	"github.com/vency7744/pmos-panel/internal/logger"
	"github.com/vency7744/pmos-panel/internal/monitor"
	"github.com/vency7744/pmos-panel/internal/power"
	"github.com/vency7744/pmos-panel/internal/service"
	"github.com/vency7744/pmos-panel/internal/terminal"
	pmosweb "github.com/vency7744/pmos-panel/web"
)

func main() {
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Logging.Level, cfg.Logging.Format)

	log.Info("starting pmos-panel",
		slog.String("listen", cfg.ListenAddr()),
		slog.String("log_level", cfg.Logging.Level),
	)

	authMgr := auth.NewManager(cfg.Auth.SecretKey, cfg.Auth.SessionTimeout, log)
	authMgr.StartSessionCleanup(5 * time.Minute)

	if os.Getenv("PMOS_PANEL_USERNAME") != "" && os.Getenv("PMOS_PANEL_PASSWORD") != "" {
		authMgr.SetUser(os.Getenv("PMOS_PANEL_USERNAME"), os.Getenv("PMOS_PANEL_PASSWORD"))
	} else if !authMgr.HasUsers() {
		authMgr.SetUser("admin", "admin1234")
		log.Warn("default user created (admin/admin1234) — change immediately")
	}

	collector := monitor.NewCollector("/")
	term := terminal.New(log)
	svcMgr := service.New(log)
	fileMgr := filemanager.New(cfg.FilePath.Root, log)
	journ := journald.New(log)
	pwr := power.New(log)

	handler := api.NewHandler(log, authMgr)
	statsWS := api.NewStatsWS(collector, log, authMgr)

	services := &api.Services{
		Terminal: term,
		Services: svcMgr,
		Files:    fileMgr,
		Journal:  journ,
		Power:    pwr,
		Auth:     authMgr,
		Logger:   log,
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	statsWS.RegisterRoutes(mux)
	services.RegisterRoutes(mux)

	mux.HandleFunc("GET /api/ws/terminal", func(w http.ResponseWriter, r *http.Request) {
		_, err := authMgr.WSAuth(r)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		term.HandleWS(w, r)
	})

	staticFS, err := fs.Sub(pmosweb.StaticFiles, ".")
	if err != nil {
		log.Warn("embedded filesystem not available, using disk fallback")
		mux.Handle("/", http.FileServer(http.Dir("web/dist")))
	} else {
		mux.Handle("/", embeddedFileHandler(staticFS))
		log.Info("serving embedded frontend")
	}

	srv := &http.Server{
		Addr:         cfg.ListenAddr(),
		Handler:      loggingMiddleware(log, mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Info("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Info("server listening", slog.String("addr", cfg.ListenAddr()))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped")
}

func loggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if r.URL.Path == "/api/ws/stats" || r.URL.Path == "/api/ws/terminal" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote", r.RemoteAddr),
			slog.Duration("duration", time.Since(start)),
		)
	})
}
