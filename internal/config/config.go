package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Auth     AuthConfig     `json:"auth"`
	Logging  LoggingConfig  `json:"logging"`
	FilePath FilePathConfig `json:"file_path"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type AuthConfig struct {
	SecretKey      string `json:"secret_key"`
	SessionTimeout int    `json:"session_timeout_minutes"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type FilePathConfig struct {
	Root string `json:"root"`
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Auth: AuthConfig{
			SecretKey:      "",
			SessionTimeout: 60,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		FilePath: FilePathConfig{
			Root: filepath.Join(home, "pmos-panel"),
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	applyEnvOverrides(cfg)

	if cfg.Auth.SecretKey == "" {
		cfg.Auth.SecretKey = generateRandomKey(32)
	}

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("PMOS_PANEL_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("PMOS_PANEL_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("PMOS_PANEL_SECRET_KEY"); v != "" {
		cfg.Auth.SecretKey = v
	}
	if v := os.Getenv("PMOS_PANEL_SESSION_TIMEOUT"); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			cfg.Auth.SessionTimeout = t
		}
	}
	if v := os.Getenv("PMOS_PANEL_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("PMOS_PANEL_FILE_ROOT"); v != "" {
		cfg.FilePath.Root = v
	}
}

func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func generateRandomKey(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%016x%016x%016x%016x", 0, 0, 0, 0)
	}
	return hex.EncodeToString(b)
}
