package handler

import (
	"net/http"
	"os"

	"sencha/backend/internal/sengen"

	"github.com/gin-gonic/gin"
)

func GetPromptHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"text": sengen.GetPrompt()})
}

func UpdatePromptHandler(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "code": "INVALID_REQUEST"})
		return
	}
	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt text is required", "code": "INVALID_PROMPT"})
		return
	}

	if err := os.WriteFile("prompt.tmpl", []byte(req.Text), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save prompt", "code": "PROMPT_SAVE_ERROR"})
		return
	}

	if err := sengen.ReloadPrompt(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reload prompt", "code": "PROMPT_RELOAD_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prompt updated"})
}
