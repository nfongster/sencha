package main

import (
	"fmt"
	"log"
	"os"

	"sencha/backend/internal/config"
	"sencha/backend/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		wd, _ := os.Getwd()
		fmt.Fprintf(os.Stderr, "Warning: could not load config.json (tried: %s/config.json): %v\n", wd, err)
		cfg = config.Defaults()
		log.Printf("[main] config load failed, using defaults — base_url=%q model=%q", cfg.LLM.BaseURL, cfg.LLM.Model)
	}

	log.Printf("[main] initializing handler with config — base_url=%q model=%q", cfg.LLM.BaseURL, cfg.LLM.Model)
	handler.Initialize(cfg)

	r := gin.Default()
	handler.RegisterRoutes(r)
	fmt.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
