package app

import (
	"fmt"
	"net"
	"time"
)

type Config struct {
	Address         string
	DataDir         string
	ShutdownTimeout time.Duration
}

func DefaultConfig() Config {
	return Config{Address: "127.0.0.1:21206", DataDir: "./data", ShutdownTimeout: 5 * time.Second}
}

func (c Config) Validate() error {
	if c.DataDir == "" {
		return fmt.Errorf("data directory is required")
	}
	host, port, err := net.SplitHostPort(c.Address)
	if err != nil {
		return fmt.Errorf("invalid HTTP address: %w", err)
	}
	if host != "127.0.0.1" && host != "localhost" {
		return fmt.Errorf("MineAir must bind to the local control interface")
	}
	if port == "" {
		return fmt.Errorf("HTTP port is required")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}
	return nil
}
