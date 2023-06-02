package config

import (
    "fmt"
    "os"
)

const (
    envClientServiceHost = "CLIENT_SERVICE_HOST"
    envClientServicePort = "CLIENT_SERVICE_PORT"
)

// ClientAppConfig ...
type ClientAppConfig struct {
    Host string
    Port string
    // Protocol gRPC or HTTP
    Protocol string
}

// GetCombinedAddress with Host and Port
func (cfg *ClientAppConfig) GetCombinedAddress() string {
    return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}

// LoadFromEnv form environment variables
func (cfg *ClientAppConfig) LoadFromEnv() {
    cfg.Host = os.Getenv(envClientServiceHost)
    if len(cfg.Host) == 0 {
        cfg.Host = "0.0.0.0"
    }
    cfg.Port = os.Getenv(envClientServicePort)
    if len(cfg.Port) == 0 {
        cfg.Port = "50051"
    }

    cfg.Protocol = "gRPC"
}

// String impl
func (cfg *ClientAppConfig) String() string {
    return fmt.Sprintf(
        "---Client Configuration---\nProtocol:%s\nHost:%s\nPort:%s\n",
        cfg.Protocol,
        cfg.Host,
        cfg.Port,
    )
}
