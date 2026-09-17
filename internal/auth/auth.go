package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionExpired    = errors.New("session expired")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrRateLimited       = errors.New("too many attempts, try again later")
)

type User struct {
	Username     string
	PasswordHash string
	Salt         string
}

type Session struct {
	Username  string
	ExpiresAt time.Time
}

type loginAttempt struct {
	count     int
	firstSeen time.Time
}

type Manager struct {
	users    map[string]*User
	sessions map[string]*Session
	attempts map[string]*loginAttempt
	mu       sync.RWMutex
	timeout  time.Duration
	secret   string
	logger   *slog.Logger
}

func NewManager(secret string, sessionTimeoutMinutes int, logger *slog.Logger) *Manager {
	return &Manager{
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
		attempts: make(map[string]*loginAttempt),
		timeout:  time.Duration(sessionTimeoutMinutes) * time.Minute,
		secret:   secret,
		logger:   logger,
	}
}

func (m *Manager) checkRateLimit(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, exists := m.attempts[ip]
	if !exists {
		m.attempts[ip] = &loginAttempt{count: 1, firstSeen: time.Now()}
		return true
	}

	if time.Since(a.firstSeen) > 5*time.Minute {
		a.count = 1
		a.firstSeen = time.Now()
		return true
	}

	if a.count >= 5 {
		return false
	}

	a.count++
	return true
}

func (m *Manager) resetRateLimit(ip string) {
	m.mu.Lock()
	delete(m.attempts, ip)
	m.mu.Unlock()
}

func (m *Manager) SetUser(username, password string) error {
	salt := generateRandomHex(16)
	hash := deriveKey(password, salt, m.secret)
	m.mu.Lock()
	m.users[username] = &User{
		Username:     username,
		PasswordHash: hash,
		Salt:         salt,
	}
	m.mu.Unlock()
	m.logger.Info("user configured", slog.String("username", username))
	return nil
}

func (m *Manager) Authenticate(username, password, clientIP string) (string, error) {
	if !m.checkRateLimit(clientIP) {
		m.logger.Warn("rate limited login attempt", slog.String("ip", clientIP))
		return "", ErrRateLimited
	}

	m.mu.RLock()
	user, exists := m.users[username]
	m.mu.RUnlock()

	if !exists {
		return "", ErrInvalidCredentials
	}

	hash := deriveKey(password, user.Salt, m.secret)
	if subtle.ConstantTimeCompare([]byte(hash), []byte(user.PasswordHash)) != 1 {
		m.logger.Warn("failed login attempt", slog.String("username", username), slog.String("ip", clientIP))
		return "", ErrInvalidCredentials
	}

	m.resetRateLimit(clientIP)

	token := generateRandomHex(32)
	m.mu.Lock()
	m.sessions[token] = &Session{
		Username:  username,
		ExpiresAt: time.Now().Add(m.timeout),
	}
	m.mu.Unlock()

	m.logger.Info("user logged in", slog.String("username", username))
	return token, nil
}

func (m *Manager) Validate(token string) (string, error) {
	m.mu.RLock()
	session, exists := m.sessions[token]
	m.mu.RUnlock()

	if !exists {
		return "", ErrUnauthorized
	}

	if time.Now().After(session.ExpiresAt) {
		m.mu.Lock()
		delete(m.sessions, token)
		m.mu.Unlock()
		return "", ErrSessionExpired
	}

	return session.Username, nil
}

func (m *Manager) Logout(token string) {
	m.mu.Lock()
	if session, exists := m.sessions[token]; exists {
		delete(m.sessions, token)
		m.logger.Info("user logged out", slog.String("username", session.Username))
	}
	m.mu.Unlock()
}

func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		username, err := m.Validate(token.Value)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-Username", username)
		next.ServeHTTP(w, r)
	})
}

func (m *Manager) WSAuth(r *http.Request) (string, error) {
	token, err := r.Cookie("session_token")
	if err != nil {
		return "", ErrUnauthorized
	}
	return m.Validate(token.Value)
}

func deriveKey(password, salt, secret string) string {
	h := sha256.New()
	io.WriteString(h, salt)
	io.WriteString(h, password)
	io.WriteString(h, secret)
	for i := 0; i < 100000; i++ {
		h2 := sha256.New()
		h2.Write(h.Sum(nil))
		h.Reset()
		h.Write(h2.Sum(nil))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Manager) SetSecret(s string) {
	m.secret = s
}

func (m *Manager) CleanupSessions() {
	m.mu.Lock()
	now := time.Now()
	for token, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, token)
		}
	}
	m.mu.Unlock()
}

func (m *Manager) StartSessionCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			m.CleanupSessions()
		}
	}()
}

func (m *Manager) UserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.users)
}

func (m *Manager) HasUsers() bool {
	return m.UserCount() > 0
}

func (m *Manager) SetupFirstUser(w http.ResponseWriter, r *http.Request) {
	if m.HasUsers() {
		http.Error(w, `{"error":"users already configured"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"username and password required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, `{"error":"password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	m.SetUser(req.Username, req.Password)

	token, _ := m.Authenticate(req.Username, req.Password, r.RemoteAddr)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(m.timeout.Seconds()),
	})

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","username":"%s"}`, req.Username)
}
