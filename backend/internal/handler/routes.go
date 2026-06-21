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

	r.GET("/api/prompt", GetPromptHandler)
	r.PUT("/api/prompt", UpdatePromptHandler)

	r.GET("/api/phases", ListPhasesHandler)
	r.POST("/api/phases", CreatePhaseHandler)
	r.GET("/api/phases/:number/levels", LevelsInPhaseHandler)
	r.PATCH("/api/phases/:number", UpdatePhaseHandler)
	r.DELETE("/api/phases/:number", DeletePhaseHandler)

	r.POST("/api/levels", CreateLevelHandler)
	r.POST("/api/levels/extract-from-url", ExtractFromUrlHandler)
	r.GET("/api/levels/categories", GetCategoriesHandler)
	r.GET("/api/levels/max", MaxLevelHandler)
	r.GET("/api/levels/:number", GetLevelHandler)
	r.PATCH("/api/levels/:number", UpdateLevelRulesHandler)
	r.PUT("/api/levels/:number/vocabulary", UpdateLevelVocabularyHandler)
	r.DELETE("/api/levels/:number", DeleteLevelHandler)

	r.GET("/api/levels/:number/sentences/count", CountSentencesHandler)
	r.GET("/api/levels/:number/sentences", ListSentencesHandler)
	r.DELETE("/api/levels/:number/sentences", DeleteSentencesHandler)
	r.POST("/api/levels/:number/sentences/generate", GenerateSentencesHandler)

	api := r.Group("/api/sessions")
	api.POST("", CreateSessionHandler)
	api.GET("/:id", GetSessionHandler)
	api.POST("/:id/reveal", RevealHandler)
	api.POST("/:id/grade", GradeHandler)
}
