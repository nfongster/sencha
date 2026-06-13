package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func startHandler(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		input := c.PostForm("input")
		c.String(http.StatusOK, "Received: %s\n", input)
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Invalid request method"})
	}
}

func main() {
	r := gin.Default()
	r.POST("/start", startHandler)
	fmt.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
