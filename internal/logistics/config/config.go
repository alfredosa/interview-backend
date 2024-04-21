package config

import (
	"errors"
	"fmt"
	"os"
)

// Might be better directly stored inside the func :)
const (
	envClientServiceHost   = "CLIENT_SERVICE_HOST"
	envClientServicePort   = "CLIENT_SERVICE_PORT"
	envClientTransportType = "CLIENT_TRANSPORT_TYPE"
	envClientHTTPScheme    = "CLIENT_HTTP_SCHEME"
)

var ErrConfigurationRequired = errors.New("configuration required: both host and port must be set")

// ServerConfig
type ServerConfig struct {
	Host string
	Port string
}

// Config is an interface for configurations that can load envs
// and provide a combined host and port address.
type Config interface {
	LoadFromEnv()
	GetCombinedAddress() string
}

// GetCombinedAddress with Host and Port
func (cfg *ServerConfig) GetCombinedAddress() string {
	return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}

// Loads the server configuration from the Env variables.
func (sc *ServerConfig) LoadFromEnv() error {
	sc.Host = os.Getenv("SERVER_HOST")
	sc.Port = os.Getenv("SERVER_TCP_PORT")

	if len(sc.Host) == 0 || len(sc.Port) == 0 {
		return ErrConfigurationRequired
	}

	return nil
}

// ClientAppConfig ...
type ClientAppConfig struct {
	Host   string
	Port   string
	Scheme string
	// TransportTypeProtocol gRPC or HTTP.
	TransportTypeProtocol string
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

	cfg.TransportTypeProtocol = os.Getenv(envClientTransportType)
	cfg.Scheme = os.Getenv(envClientHTTPScheme)
}

// String impl
func (cfg *ClientAppConfig) String() string {
	return fmt.Sprintf(
		"---Client Configuration---\nProtocol:%s\nHost:%s\nPort:%s\n",
		cfg.TransportTypeProtocol,

		cfg.Host,
		cfg.Port,
	)
}
