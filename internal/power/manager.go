package power

import (
	"fmt"
	"log/slog"
	"os/exec"
	"os/user"
)

type Manager struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Manager {
	return &Manager{logger: logger}
}

func isRoot() bool {
	u, err := user.Current()
	if err != nil {
		return false
	}
	return u.Uid == "0"
}

func (m *Manager) Reboot() error {
	m.logger.Warn("system reboot requested")
	if isRoot() {
		return exec.Command("reboot").Run()
	}
	return exec.Command("sudo", "reboot").Run()
}

func (m *Manager) Shutdown() error {
	m.logger.Warn("system shutdown requested")
	if isRoot() {
		return exec.Command("poweroff").Run()
	}
	return exec.Command("sudo", "poweroff").Run()
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
