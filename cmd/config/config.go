package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	defaultHost     = ":8080"
	defaultLogLevel = "DEBUG"
	defaultDBDsn    = "postgresql://localhost:5432/melifaro"
	//defaultDBDsn = ""
)

type Configuration struct {
	ServeAddress string
	LogLevel     string
	DBDsn        string
}

func Load() (*Configuration, error) {
	var cfg Configuration
	flag.StringVar(&cfg.ServeAddress, "a", defaultHost, "Address to listen on")
	flag.StringVar(&cfg.LogLevel, "l", defaultLogLevel, "Log level")
	flag.StringVar(&cfg.DBDsn, "d", defaultDBDsn, "Database dsn")
	flag.Parse()
	if serverAddr, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		cfg.ServeAddress = serverAddr
	}
	if logLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = logLevel
	}
	if dbDsn, ok := os.LookupEnv("DATABASE_URI"); ok {
		cfg.DBDsn = dbDsn
	}
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Configuration) Validate() error {
	if err := c.validateServeAddress(); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) validateServeAddress() error {
	addr := c.ServeAddress
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address: %s", addr)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %s", portStr)
	}

	if host != "localhost" && host != "" && net.ParseIP(host) == nil {
		return fmt.Errorf("invalid host: %s", host)
	}

	return nil
}
