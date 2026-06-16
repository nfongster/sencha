package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func rootHandler(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		input := c.PostForm("input")
		
		switch input {
		case "Start":
			c.String(http.StatusOK, "Starting the app...")
		case "Quit":
			c.String(http.StatusOK, "Quitting the app. Goodbye!")
		default:
			c.String(http.StatusBadRequest, fmt.Sprintf("Invalid string '%s'. Please try again.", input))
		}
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Invalid request method"})
	}
}

// The following terminal command sends strings to the server (YES, you need "input=" since we don't have dedicated JSON structs yet...):
// curl -X POST http://localhost:8080/ -d "input=Start"
func main() {
	r := gin.Default()
	r.POST("/", rootHandler)
	fmt.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
