package service

import (
	"bufio"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

type Manager struct {
	logger *slog.Logger
}

type Service struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Active string `json:"active"`
	Loaded string `json:"loaded"`
}

func New(logger *slog.Logger) *Manager {
	return &Manager{logger: logger}
}

func (m *Manager) List() ([]Service, error) {
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

func (m *Manager) Status(name string) (string, error) {
	if err := validateServiceName(name); err != nil {
		return "", err
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
	return exec.Command("systemctl", "start", name).Run()
}

func (m *Manager) Stop(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	m.logger.Info("stopping service", slog.String("name", name))
	return exec.Command("systemctl", "stop", name).Run()
}

func (m *Manager) Restart(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	m.logger.Info("restarting service", slog.String("name", name))
	return exec.Command("systemctl", "restart", name).Run()
}

func (m *Manager) Enable(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
	}
	return exec.Command("systemctl", "enable", name).Run()
}

func (m *Manager) Disable(name string) error {
	if err := validateServiceName(name); err != nil {
		return err
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
