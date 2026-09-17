package api

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/vency7744/pmos-panel/internal/auth"
	"github.com/vency7744/pmos-panel/internal/filemanager"
	"github.com/vency7744/pmos-panel/internal/journald"
	"github.com/vency7744/pmos-panel/internal/netinfo"
	"github.com/vency7744/pmos-panel/internal/power"
	"github.com/vency7744/pmos-panel/internal/process"
	"github.com/vency7744/pmos-panel/internal/service"
	"github.com/vency7744/pmos-panel/internal/terminal"
)

type Services struct {
	Terminal   *terminal.Terminal
	Services   *service.Manager
	Files      *filemanager.Manager
	Journal    *journald.Manager
	Power      *power.Manager
	Auth       *auth.Manager
	Logger     *slog.Logger
}

func (s *Services) RegisterRoutes(mux *http.ServeMux) {
	// Systemd
	mux.HandleFunc("GET /api/services", s.wrapAuth(s.handleServices))
	mux.HandleFunc("POST /api/services/{name}/start", s.wrapAuth(s.handleServiceStart))
	mux.HandleFunc("POST /api/services/{name}/stop", s.wrapAuth(s.handleServiceStop))
	mux.HandleFunc("POST /api/services/{name}/restart", s.wrapAuth(s.handleServiceRestart))
	mux.HandleFunc("POST /api/services/{name}/enable", s.wrapAuth(s.handleServiceEnable))
	mux.HandleFunc("POST /api/services/{name}/disable", s.wrapAuth(s.handleServiceDisable))

	// Processes
	mux.HandleFunc("GET /api/processes", s.wrapAuth(s.handleProcesses))
	mux.HandleFunc("POST /api/processes/{pid}/kill", s.wrapAuth(s.handleProcessKill))

	// Files
	mux.HandleFunc("GET /api/files", s.wrapAuth(s.handleFilesList))
	mux.HandleFunc("GET /api/files/read", s.wrapAuth(s.handleFilesRead))
	mux.HandleFunc("GET /api/files/download", s.wrapAuth(s.handleFilesDownload))
	mux.HandleFunc("POST /api/files/write", s.wrapAuth(s.handleFilesWrite))
	mux.HandleFunc("POST /api/files/mkdir", s.wrapAuth(s.handleFilesMkdir))
	mux.HandleFunc("POST /api/files/delete", s.wrapAuth(s.handleFilesDelete))
	mux.HandleFunc("POST /api/files/rename", s.wrapAuth(s.handleFilesRename))
	mux.HandleFunc("POST /api/files/upload", s.wrapAuth(s.handleFilesUpload))

	// Logs
	mux.HandleFunc("GET /api/logs", s.wrapAuth(s.handleLogs))
	mux.HandleFunc("GET /api/logs/services", s.wrapAuth(s.handleLogServices))
	mux.HandleFunc("GET /api/ws/logs", s.handleLogsWS)

	// Network
	mux.HandleFunc("GET /api/network", s.wrapAuth(s.handleNetwork))

	// Power
	mux.HandleFunc("POST /api/power/reboot", s.wrapAuth(s.handleReboot))
	mux.HandleFunc("POST /api/power/shutdown", s.wrapAuth(s.handleShutdown))
}

func (s *Services) wrapAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		_, err = s.Auth.Validate(token.Value)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Services) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Systemd handlers
func (s *Services) handleServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.Services.List()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, services)
}

func (s *Services) handleServiceStart(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.Services.Start(name); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Services) handleServiceStop(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.Services.Stop(name); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Services) handleServiceRestart(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.Services.Restart(name); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}

func (s *Services) handleServiceEnable(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.Services.Enable(name); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

func (s *Services) handleServiceDisable(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.Services.Disable(name); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

// Process handlers
func (s *Services) handleProcesses(w http.ResponseWriter, r *http.Request) {
	procs, err := process.List()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, procs)
}

func (s *Services) handleProcessKill(w http.ResponseWriter, r *http.Request) {
	pidStr := r.PathValue("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid pid"})
		return
	}
	if err := process.Kill(pid); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "killed"})
}

// File handlers
func (s *Services) handleFilesList(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("path")
	if dir == "" {
		dir = "/"
	}
	files, err := s.Files.List(dir)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, files)
}

func (s *Services) handleFilesRead(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path required"})
		return
	}
	if err := s.Files.ServeRead(w, r, path); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func (s *Services) handleFilesDownload(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path required"})
		return
	}
	if err := s.Files.ServeDownload(w, r, path); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}

func (s *Services) handleFilesWrite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := s.Files.Write(req.Path, []byte(req.Content)); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Services) handleFilesMkdir(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := s.Files.Mkdir(req.Path); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Services) handleFilesDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := s.Files.Delete(req.Path); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Services) handleFilesRename(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := s.Files.Rename(req.OldPath, req.NewPath); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Services) handleFilesUpload(w http.ResponseWriter, r *http.Request) {
	dir := r.FormValue("path")
	if dir == "" {
		dir = "/"
	}
	files, err := s.Files.Upload(r, dir)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "files": files})
}

// Log handlers
func (s *Services) handleLogs(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	query := r.URL.Query().Get("query")
	linesStr := r.URL.Query().Get("lines")
	lines := 100
	if linesStr != "" {
		if n, err := strconv.Atoi(linesStr); err == nil && n > 0 && n <= 1000 {
			lines = n
		}
	}

	var entries []journald.LogEntry
	var err error

	if query != "" {
		entries, err = s.Journal.Search(query, serviceName, lines)
	} else {
		entries, err = s.Journal.Recent(serviceName, lines)
	}

	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, entries)
}

func (s *Services) handleLogServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.Journal.Services()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, services)
}

var logWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (s *Services) handleLogsWS(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if _, err = s.Auth.Validate(token.Value); err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	conn, err := logWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Logger.Error("ws upgrade failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	serviceName := r.URL.Query().Get("service")

	pr, pw := io.Pipe()
	closer, err := s.Journal.Follow(serviceName, pw)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"failed to start log stream"}`))
		return
	}
	defer closer.Close()
	pw.Close()

	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			entry := journald.LogEntry{Message: line}
			if len(line) > 19 {
				entry.Timestamp = line[:19]
				entry.Message = strings.TrimSpace(line[19:])
			}
			data, _ := json.Marshal(entry)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

// Network handler
func (s *Services) handleNetwork(w http.ResponseWriter, r *http.Request) {
	ifaces, err := netinfo.List()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, ifaces)
}

// Power handlers
func (s *Services) handleReboot(w http.ResponseWriter, r *http.Request) {
	if err := s.Power.Reboot(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "rebooting"})
}

func (s *Services) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if err := s.Power.Shutdown(); err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "shutting down"})
}


