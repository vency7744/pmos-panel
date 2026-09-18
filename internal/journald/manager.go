package journald

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type Manager struct {
	logger     *slog.Logger
	useOpenRC  bool
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Priority  string `json:"priority"`
	Service   string `json:"service"`
	Message   string `json:"message"`
}

func New(logger *slog.Logger) *Manager {
	useOpenRC := false
	if _, err := exec.LookPath("journalctl"); err != nil {
		useOpenRC = true
		logger.Info("using OpenRC log system (logread)")
	} else {
		logger.Info("using systemd log system (journalctl)")
	}
	return &Manager{logger: logger, useOpenRC: useOpenRC}
}

func (m *Manager) Recent(service string, lines int) ([]LogEntry, error) {
	if m.useOpenRC {
		return m.recentOpenRC(service, lines)
	}
	return m.recentSystemd(service, lines)
}

func (m *Manager) recentSystemd(service string, lines int) ([]LogEntry, error) {
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

func (m *Manager) recentOpenRC(service string, lines int) ([]LogEntry, error) {
	args := []string{"-n", fmt.Sprintf("%d", lines)}
	cmd := exec.Command("logread", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("logread: %w", err)
	}

	entries := parseLogReadOutput(string(out))
	if service != "" {
		var filtered []LogEntry
		for _, e := range entries {
			if strings.Contains(e.Service, service) || strings.Contains(e.Message, service) {
				filtered = append(filtered, e)
			}
		}
		return filtered, nil
	}
	return entries, nil
}

func (m *Manager) Search(query string, service string, lines int) ([]LogEntry, error) {
	if m.useOpenRC {
		return m.searchOpenRC(query, service, lines)
	}
	return m.searchSystemd(query, service, lines)
}

func (m *Manager) searchSystemd(query string, service string, lines int) ([]LogEntry, error) {
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

func (m *Manager) searchOpenRC(query string, service string, lines int) ([]LogEntry, error) {
	args := []string{"-n", "1000"}
	cmd := exec.Command("logread", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("logread: %w", err)
	}

	entries := parseLogReadOutput(string(out))
	query = strings.ToLower(query)
	service = strings.ToLower(service)
	var filtered []LogEntry
	for _, e := range entries {
		if query != "" && !strings.Contains(strings.ToLower(e.Message), query) {
			continue
		}
		if service != "" && !strings.Contains(strings.ToLower(e.Service), service) && !strings.Contains(strings.ToLower(e.Message), service) {
			continue
		}
		filtered = append(filtered, e)
		if len(filtered) >= lines {
			break
		}
	}
	return filtered, nil
}

func (m *Manager) Services() ([]string, error) {
	if m.useOpenRC {
		return m.servicesOpenRC()
	}
	return m.servicesSystemd()
}

func (m *Manager) servicesSystemd() ([]string, error) {
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

func (m *Manager) servicesOpenRC() ([]string, error) {
	entries, err := os.ReadDir("/etc/init.d")
	if err != nil {
		return nil, err
	}
	var services []string
	for _, e := range entries {
		if !e.IsDir() {
			services = append(services, e.Name())
		}
	}
	return services, nil
}

func (m *Manager) Follow(service string, writer io.Writer) (io.Closer, error) {
	if m.useOpenRC {
		return m.followOpenRC(service, writer)
	}
	return m.followSystemd(service, writer)
}

func (m *Manager) followSystemd(service string, writer io.Writer) (io.Closer, error) {
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
	return &logCloser{cmd: cmd}, nil
}

func (m *Manager) followOpenRC(service string, writer io.Writer) (io.Closer, error) {
	cmd := exec.Command("tail", "-f", "-n", "0", "/var/log/messages")
	if service != "" {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("logread -f | grep --line-buffered '%s'", service))
	}
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start log follower: %w", err)
	}
	return &logCloser{cmd: cmd}, nil
}

type logCloser struct {
	cmd *exec.Cmd
}

func (l *logCloser) Close() error {
	return l.cmd.Process.Kill()
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

func parseLogReadOutput(output string) []LogEntry {
	var entries []LogEntry
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		entry := LogEntry{Message: line}
		// Format: "Sep 18 19:13:15 hostname service[pid]: message"
		if idx := strings.Index(line, "] "); idx > 0 {
			msgPart := line[idx+2:]
			servicePart := line[:idx]
			if lastSlash := strings.LastIndex(servicePart, " "); lastSlash > 0 {
				service := servicePart[lastSlash+1:]
				if parenStart := strings.Index(service, "["); parenStart > 0 {
					service = service[:parenStart]
				}
				entry.Service = service
			}
			entry.Message = msgPart
		}
		if len(line) > 15 {
			entry.Timestamp = line[:15]
		}
		entries = append(entries, entry)
	}
	return entries
}
