package main

import (
	"log"

	"github.com/coopnorge/interview-backend/internal/logistics/config"
	"github.com/coopnorge/interview-backend/internal/logistics/services/server"
)

func main() {
	cfg := &config.ServerConfig{}
	cfg.LoadFromEnv()

	log.Println("Loaded Configuration from Environment Variables\n", cfg)
	server.ListendAndAccept(cfg)
}
