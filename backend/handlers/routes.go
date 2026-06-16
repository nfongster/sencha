package handlers

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	r.GET("/api/health", HealthHandler)

	api := r.Group("/api/sessions")
	api.POST("", CreateSessionHandler)
	api.GET("/:id", GetSessionHandler)
	api.POST("/:id/reveal", RevealHandler)
	api.POST("/:id/grade", GradeHandler)
}
