package main

import (
	"fmt"
	"os"

	"sencha/backend/internal/config"
	"sencha/backend/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config.json: %v\n", err)
		cfg = config.Defaults()
	}

	handler.Initialize(cfg)

	r := gin.Default()
	handler.RegisterRoutes(r)
	fmt.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
