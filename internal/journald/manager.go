package journald

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
)

type Manager struct {
	logger *slog.Logger
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Priority  string `json:"priority"`
	Service   string `json:"service"`
	Message   string `json:"message"`
}

func New(logger *slog.Logger) *Manager {
	return &Manager{logger: logger}
}

func (m *Manager) Recent(service string, lines int) ([]LogEntry, error) {
	args := []string{"--no-pager", "--output=short-iso", "-n", fmt.Sprintf("%d", lines)}
	if service != "" {
		args = append(args, "-u", service)
	}
	cmd := exec.Command("journalctl", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("journalctl: %w", err)
	}
	return parseJournalOutput(string(out)), nil
}

func (m *Manager) Search(query string, service string, lines int) ([]LogEntry, error) {
	args := []string{"--no-pager", "--output=short-iso", "-n", fmt.Sprintf("%d", lines)}
	if query != "" {
		args = append(args, "-g", query)
	}
	if service != "" {
		args = append(args, "-u", service)
	}
	cmd := exec.Command("journalctl", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("journalctl: %w", err)
	}
	return parseJournalOutput(string(out)), nil
}

func (m *Manager) Services() ([]string, error) {
	cmd := exec.Command("journalctl", "--no-pager", "--field=_SYSTEMD_UNIT")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var services []string
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		svc := strings.TrimSpace(scanner.Text())
		if svc != "" && !seen[svc] {
			seen[svc] = true
			services = append(services, svc)
		}
	}
	return services, nil
}

func parseJournalOutput(output string) []LogEntry {
	var entries []LogEntry
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		entry := LogEntry{Message: line}
		if len(line) > 19 {
			entry.Timestamp = line[:19]
			rest := strings.TrimSpace(line[19:])
			entry.Message = rest
		}
		entries = append(entries, entry)
	}
	return entries
}

func (m *Manager) Follow(service string, writer io.Writer) (io.Closer, error) {
	args := []string{"--no-pager", "--output=short-iso", "-f", "-n", "0"}
	if service != "" {
		args = append(args, "-u", service)
	}
	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start journalctl: %w", err)
	}
	return &journalCloser{cmd: cmd}, nil
}

type journalCloser struct {
	cmd *exec.Cmd
}

func (j *journalCloser) Close() error {
	return j.cmd.Process.Kill()
}
