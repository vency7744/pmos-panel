package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vency7744/pmos-panel/internal/auth"
	"github.com/vency7744/pmos-panel/internal/monitor"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		return u.Host == r.Host
	},
}

type StatsWS struct {
	collector *monitor.Collector
	logger    *slog.Logger
	auth      *auth.Manager
}

func NewStatsWS(collector *monitor.Collector, logger *slog.Logger, authMgr *auth.Manager) *StatsWS {
	return &StatsWS{
		collector: collector,
		logger:    logger,
		auth:      authMgr,
	}
}

func (s *StatsWS) HandleWS(w http.ResponseWriter, r *http.Request) {
	_, err := s.auth.WSAuth(r)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("websocket upgrade failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	s.logger.Info("websocket client connected")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats, err := s.collector.Collect()
			if err != nil {
				s.logger.Error("collect stats failed", slog.String("error", err.Error()))
				continue
			}

			data, err := json.Marshal(stats)
			if err != nil {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				s.logger.Info("websocket client disconnected")
				return
			}

		case <-r.Context().Done():
			return
		}
	}
}

func (s *StatsWS) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ws/stats", s.HandleWS)
	mux.HandleFunc("GET /api/stats", s.HandleStatsAPIAuth)
}

func (s *StatsWS) HandleStatsAPIAuth(w http.ResponseWriter, r *http.Request) {
	_, err := s.auth.WSAuth(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	s.HandleStatsAPI(w, r)
}

func (s *StatsWS) HandleStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats, err := s.collector.Collect()
	if err != nil {
		http.Error(w, `{"error":"collect failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
