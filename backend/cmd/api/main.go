package main

import (
	"fmt"
	"log"
	"os"

	"sencha/backend/internal/config"
	"sencha/backend/internal/handler"

	"github.com/gin-gonic/gin"
)

func loadConfig(path string) (*config.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("could not load config from %s/%s: %w", wd, path, err)
	}
	return cfg, nil
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("fatal: %v\n\nMake sure you run from the project root, e.g.:\n\tgo run ./cmd/api/\n", err)
	}

	handler.Initialize(cfg)

	r := gin.Default()
	handler.RegisterRoutes(r)
	fmt.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
