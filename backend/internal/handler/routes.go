package handler

import (
	"sencha/backend/internal/config"
	"sencha/backend/internal/sengen"

	"github.com/gin-gonic/gin"
)

var appConfig *config.Config

func Initialize(cfg *config.Config) {
	appConfig = cfg
	sengen.Init(&cfg.LLM)
}

func RegisterRoutes(r *gin.Engine) {
	r.GET("/api/health", HealthHandler)

	api := r.Group("/api/sessions")
	api.POST("", CreateSessionHandler)
	api.GET("/:id", GetSessionHandler)
	api.POST("/:id/reveal", RevealHandler)
	api.POST("/:id/grade", GradeHandler)
}
