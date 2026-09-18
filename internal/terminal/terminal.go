package terminal

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
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

type Terminal struct {
	logger *slog.Logger
}

type WSMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

func New(logger *slog.Logger) *Terminal {
	return &Terminal{logger: logger}
}

func (t *Terminal) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		t.logger.Error("terminal ws upgrade failed", slog.String("error", err.Error()))
		return
	}
	defer conn.Close()

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	home := "/home/fw"
	user := "fw"

	cmd := exec.Command(shell)
	cmd.Dir = home

	var env []string
	for _, e := range os.Environ() {
		k := e[:strings.IndexByte(e, '=')]
		if k == "HOME" || k == "USER" || k == "SHELL" || k == "LOGNAME" {
			continue
		}
		env = append(env, e)
	}
	env = append(env,
		"TERM=xterm-256color",
		"HOME="+home,
		"USER="+user,
		"LOGNAME="+user,
		"SHELL="+shell,
	)
	cmd.Env = env

	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.logger.Error("failed to start pty", slog.String("error", err.Error()))
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","data":"failed to start shell"}`))
		return
	}
	defer func() {
		ptmx.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	t.logger.Info("terminal session started", slog.Int("pid", cmd.Process.Pid))

	var mu sync.Mutex

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			mu.Lock()
			conn.WriteMessage(websocket.TextMessage, buf[:n])
			mu.Unlock()
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "input":
			io.WriteString(ptmx, wsMsg.Data)
		case "resize":
			if wsMsg.Cols > 0 && wsMsg.Rows > 0 {
				pty.Setsize(ptmx, &pty.Winsize{
					Cols: wsMsg.Cols,
					Rows: wsMsg.Rows,
				})
			}
		case "ping":
			mu.Lock()
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong"}`))
			mu.Unlock()
		}
	}

	t.logger.Info("terminal session ended", slog.Int("pid", cmd.Process.Pid))
}
