package main

import (
	"fmt"

	"sencha/backend/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	handler.RegisterRoutes(r)
	fmt.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
