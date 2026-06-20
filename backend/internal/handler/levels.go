package handler

import (
	"net/http"
	"strconv"

	"sencha/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func MaxLevelHandler(c *gin.Context) {
	max, err := appConfig.Repository.MaxLevelNumber()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to get max level number",
			Code:  "LEVEL_NUMBER_ERROR",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"max": max})
}

func GetLevelHandler(c *gin.Context) {
	numberStr := c.Param("number")
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid level number",
			Code:  "INVALID_LEVEL_NUMBER",
		})
		return
	}

	level, err := appConfig.Repository.Level(number)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "level not found",
			Code:  "LEVEL_NOT_FOUND",
		})
		return
	}

	vocab, err := appConfig.Repository.VocabularyUpTo(number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to fetch vocabulary",
			Code:  "VOCAB_ERROR",
		})
		return
	}

	levelVocab, err := appConfig.Repository.VocabularyForLevel(number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to fetch level vocabulary",
			Code:  "VOCAB_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"level":            level,
		"vocabulary":       vocab,
		"level_vocabulary": levelVocab,
	})
}

type createLevelVocabularyItem struct {
	Korean  string `json:"korean"`
	English string `json:"english"`
}

type createLevelRequest struct {
	PhaseNumber int                         `json:"phase_number"`
	GrammarMD   string                      `json:"grammar_markdown"`
	Exceptions  string                      `json:"exceptions_markdown"`
	Vocabulary  []createLevelVocabularyItem `json:"vocabulary"`
}

func CreateLevelHandler(c *gin.Context) {
	var req createLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.GrammarMD == "" {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "grammar_markdown is required",
			Code:  "MISSING_GRAMMAR",
		})
		return
	}

	maxLevel, err := appConfig.Repository.MaxLevelNumber()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to check level number",
			Code:  "LEVEL_NUMBER_ERROR",
		})
		return
	}

	nextLevel := maxLevel + 1

	vocab := make([]repository.VocabEntry, len(req.Vocabulary))
	for i, v := range req.Vocabulary {
		vocab[i] = repository.VocabEntry{
			Korean:  v.Korean,
			English: v.English,
		}
	}

	if err := appConfig.Repository.CreateLevel(repository.Level{
		Number:       nextLevel,
		PhaseNumber:  req.PhaseNumber,
		GrammarMD:    req.GrammarMD,
		ExceptionsMD: req.Exceptions,
	}); err != nil {
		c.JSON(http.StatusConflict, errorResponse{
			Error: err.Error(),
			Code:  "LEVEL_CREATE_ERROR",
		})
		return
	}

	if len(vocab) > 0 {
		if err := appConfig.Repository.AddVocabulary(nextLevel, vocab); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				Error: "level created but failed to add vocabulary",
				Code:  "VOCAB_ADD_ERROR",
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "level created",
		"level_number": nextLevel,
	})
}

type updateLevelRulesRequest struct {
	GrammarMD  string `json:"grammar_markdown"`
	Exceptions string `json:"exceptions_markdown"`
}

func UpdateLevelRulesHandler(c *gin.Context) {
	numberStr := c.Param("number")
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid level number",
			Code:  "INVALID_LEVEL_NUMBER",
		})
		return
	}

	var req updateLevelRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.GrammarMD == "" && req.Exceptions == "" {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "at least one of grammar_markdown or exceptions_markdown must be provided",
			Code:  "MISSING_FIELDS",
		})
		return
	}

	if _, err := appConfig.Repository.Level(number); err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "level not found",
			Code:  "LEVEL_NOT_FOUND",
		})
		return
	}

	if err := appConfig.Repository.UpdateLevel(number, req.GrammarMD, req.Exceptions); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: err.Error(),
			Code:  "LEVEL_UPDATE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "level rules updated"})
}

func DeleteLevelHandler(c *gin.Context) {
	numberStr := c.Param("number")
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid level number",
			Code:  "INVALID_LEVEL_NUMBER",
		})
		return
	}

	if err := appConfig.Repository.DeleteLevel(number); err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: err.Error(),
			Code:  "LEVEL_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "level deleted"})
}

type updateVocabularyRequest struct {
	Vocabulary []createLevelVocabularyItem `json:"vocabulary"`
}

func UpdateLevelVocabularyHandler(c *gin.Context) {
	numberStr := c.Param("number")
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid level number",
			Code:  "INVALID_LEVEL_NUMBER",
		})
		return
	}

	var req updateVocabularyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if _, err := appConfig.Repository.Level(number); err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "level not found",
			Code:  "LEVEL_NOT_FOUND",
		})
		return
	}

	entries := make([]repository.VocabEntry, len(req.Vocabulary))
	for i, v := range req.Vocabulary {
		entries[i] = repository.VocabEntry{
			Korean:  v.Korean,
			English: v.English,
		}
	}

	if err := appConfig.Repository.SetVocabulary(number, entries); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to update vocabulary",
			Code:  "VOCAB_UPDATE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "vocabulary updated"})
}
