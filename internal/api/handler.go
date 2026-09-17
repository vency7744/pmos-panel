package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/vency7744/pmos-panel/internal/auth"
)

type Handler struct {
	logger *slog.Logger
	auth   *auth.Manager
}

func NewHandler(logger *slog.Logger, authMgr *auth.Manager) *Handler {
	return &Handler{
		logger: logger,
		auth:   authMgr,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", h.handleLogin)
	mux.HandleFunc("POST /api/logout", h.handleLogout)
	mux.HandleFunc("GET /api/me", h.handleMe)
	mux.HandleFunc("GET /api/health", h.wrapAuth(h.handleHealth))
}

func (h *Handler) wrapAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_, err = h.auth.Validate(token.Value)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		clientIP = realIP
	}

	token, err := h.auth.Authenticate(req.Username, req.Password, clientIP)
	if err != nil {
		if err == auth.ErrRateLimited {
			h.writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many attempts, try again later"})
			return
		}
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600,
	})

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		h.auth.Logout(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	username, err := h.auth.Validate(cookie.Value)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"username": username})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
