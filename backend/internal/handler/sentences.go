package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"sencha/backend/internal/repository"
	"sencha/backend/internal/sengen"
	"sencha/backend/internal/session"

	"github.com/gin-gonic/gin"
)

type sentenceJSON struct {
	Korean  string `json:"korean"`
	English string `json:"english"`
}

func CountSentencesHandler(c *gin.Context) {
	levelNum, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid level number", Code: "INVALID_LEVEL_NUMBER"})
		return
	}
	count, err := appConfig.Repository.CountSentencesForLevel(levelNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "failed to count sentences", Code: "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func ListSentencesHandler(c *gin.Context) {
	levelNum, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid level number", Code: "INVALID_LEVEL_NUMBER"})
		return
	}
	sentences, err := appConfig.Repository.SentencesForLevel(levelNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "failed to list sentences", Code: "INTERNAL_ERROR"})
		return
	}
	js := make([]sentenceJSON, len(sentences))
	for i, s := range sentences {
		js[i] = sentenceJSON{Korean: s.Korean, English: s.English}
	}
	c.JSON(http.StatusOK, gin.H{"sentences": js})
}

func DeleteSentencesHandler(c *gin.Context) {
	levelNum, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid level number", Code: "INVALID_LEVEL_NUMBER"})
		return
	}
	if err := appConfig.Repository.DeleteSentencesForLevel(levelNum); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "failed to delete sentences", Code: "INTERNAL_ERROR"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sentences deleted"})
}

type generateSentencesRequest struct {
	Count int `json:"count"`
}

func GenerateSentencesHandler(c *gin.Context) {
	levelNum, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid level number", Code: "INVALID_LEVEL_NUMBER"})
		return
	}

	var req generateSentencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid request body", Code: "INVALID_REQUEST"})
		return
	}
	if req.Count < 1 || req.Count > 100 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "count must be between 1 and 100", Code: "INVALID_COUNT"})
		return
	}

	levelData, err := appConfig.Repository.LoadLevelData(levelNum)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: fmt.Sprintf("invalid level %d: %v", levelNum, err),
			Code:  "INVALID_LEVEL",
		})
		return
	}

	pairs, err := sengen.Generate(req.Count, *levelData)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse{
			Error: "sentence generation failed",
			Code:  "GENERATION_FAILED",
		})
		return
	}

	newSentences := sessionsToSentences(pairs, levelNum)
	if err := appConfig.Repository.SaveSentences(newSentences); err != nil {
		log.Printf("[handler] failed to save generated sentences: %v", err)
	}

	js := make([]sentenceJSON, len(newSentences))
	for i, s := range newSentences {
		js[i] = sentenceJSON{Korean: s.Korean, English: s.English}
	}
	c.JSON(http.StatusCreated, gin.H{"sentences": js})
}

func sentencesToPairs(sentences []repository.Sentence) []session.SentencePair {
	pairs := make([]session.SentencePair, len(sentences))
	for i, s := range sentences {
		pairs[i] = session.SentencePair{Korean: s.Korean, English: s.English}
	}
	return pairs
}
