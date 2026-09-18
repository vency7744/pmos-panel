package service

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type Manager struct {
	logger  *slog.Logger
	useOpenRC bool
}

type Service struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Active string `json:"active"`
	Loaded string `json:"loaded"`
}

func New(logger *slog.Logger) *Manager {
	useOpenRC := false
	if _, err := exec.LookPath("rc-service"); err == nil {
		useOpenRC = true
		logger.Info("detected OpenRC init system")
	} else {
		logger.Info("detected systemd init system")
	}
	return &Manager{logger: logger, useOpenRC: useOpenRC}
}

func (m *Manager) List() ([]Service, error) {
	if m.useOpenRC {
		return m.listOpenRC()
	}
	return m.listSystemd()
}

func (m *Manager) listSystemd() ([]Service, error) {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend", "--plain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("systemctl list-units: %w", err)
	}

	var services []Service
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if len(line) < 4 {
			continue
		}
		svc := Service{
			Name:   strings.TrimSuffix(line[0], ".service"),
			Active: line[1],
			Loaded: line[2],
			Status: line[3],
		}
		services = append(services, svc)
	}
	return services, nil
}

func (m *Manager) listOpenRC() ([]Service, error) {
	entries, err := os.ReadDir("/etc/init.d")
	if err != nil {
		return nil, fmt.Errorf("read /etc/init.d: %w", err)
	}

	var services []Service
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		status := "stopped"
		cmd := exec.Command("rc-service", name, "status")
		if err := cmd.Run(); err == nil {
			status = "started"
		}
		services = append(services, Service{
			Name:   name,
			Status: status,
			Active: status,
			Loaded: "static",
		})
	}
	return services, nil
}

func (m *Manager) Status(name string) (string, error) {
	if err := validateServiceName(name); err != nil {
		return "", err
	}
	if m.useOpenRC {
		cmd := exec.Command("rc-service", name, "status")
		if err := cmd.Run(); err != nil {
			return "stopped", nil
		}
		return "started", nil
	}
	cmd := exec.Command("systemctl", "is-active", name)
	out, err := cmd.Output()
	if err != nil {
		return "inactive", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) Start(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	m.logger.Info("starting service", slog.String("name", name))
	if m.useOpenRC {
		return exec.Command("rc-service", name, "start").Run()
	}
	return exec.Command("systemctl", "start", name).Run()
}

func (m *Manager) Stop(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	m.logger.Info("stopping service", slog.String("name", name))
	if m.useOpenRC {
		return exec.Command("rc-service", name, "stop").Run()
	}
	return exec.Command("systemctl", "stop", name).Run()
}

func (m *Manager) Restart(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	m.logger.Info("restarting service", slog.String("name", name))
	if m.useOpenRC {
		if err := exec.Command("rc-service", name, "stop").Run(); err != nil {
			m.logger.Warn("stop failed, trying start", slog.String("error", err.Error()))
		}
		return exec.Command("rc-service", name, "start").Run()
	}
	return exec.Command("systemctl", "restart", name).Run()
}

func (m *Manager) Enable(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	if m.useOpenRC {
		return exec.Command("rc-update", "add", name, "default").Run()
	}
	return exec.Command("systemctl", "enable", name).Run()
}

func (m *Manager) Disable(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	if m.useOpenRC {
		return exec.Command("rc-update", "delete", name, "default").Run()
	}
	return exec.Command("systemctl", "disable", name).Run()
}

func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name required")
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return fmt.Errorf("invalid service name")
		}
	}
	return nil
}
