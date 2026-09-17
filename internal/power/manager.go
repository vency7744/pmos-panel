package power

import (
	"fmt"
	"log/slog"
	"os/exec"
)

type Manager struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Manager {
	return &Manager{logger: logger}
}

func (m *Manager) Reboot() error {
	m.logger.Warn("system reboot requested")
	return exec.Command("systemctl", "reboot").Run()
}

func (m *Manager) Shutdown() error {
	m.logger.Warn("system shutdown requested")
	return exec.Command("systemctl", "poweroff").Run()
}

func (m *Manager) RebootScheduled(delay string) error {
	if err := validateDelay(delay); err != nil {
		return err
	}
	m.logger.Warn("scheduled reboot", slog.String("delay", delay))
	return exec.Command("shutdown", "-r", delay).Run()
}

func (m *Manager) ShutdownScheduled(delay string) error {
	if err := validateDelay(delay); err != nil {
		return err
	}
	m.logger.Warn("scheduled shutdown", slog.String("delay", delay))
	return exec.Command("shutdown", "-h", delay).Run()
}

func (m *Manager) CancelSchedule() error {
	return exec.Command("shutdown", "-c").Run()
}

func validateDelay(delay string) error {
	if delay == "" {
		return fmt.Errorf("delay required")
	}
	valid := map[string]bool{
		"now": true, "+0": true,
	}
	if valid[delay] {
		return nil
	}
	if len(delay) > 1 && delay[0] == '+' {
		for _, c := range delay[1:] {
			if c < '0' || c > '9' {
				return fmt.Errorf("invalid delay format")
			}
		}
		return nil
	}
	return fmt.Errorf("invalid delay format")
}
