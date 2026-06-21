package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"sencha/backend/internal/repository"
	"sencha/backend/internal/sengen"

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
	Korean   string `json:"korean"`
	English  string `json:"english"`
	Category string `json:"category"`
}

type createLevelRequest struct {
	PhaseNumber int                         `json:"phase_number"`
	GrammarMD   string                      `json:"grammar_markdown"`
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
			Korean:   v.Korean,
			English:  v.English,
			Category: v.Category,
		}
	}

	if err := appConfig.Repository.CreateLevel(repository.Level{
		Number:      nextLevel,
		PhaseNumber: req.PhaseNumber,
		GrammarMD:   req.GrammarMD,
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
	GrammarMD string `json:"grammar_markdown"`
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

	if req.GrammarMD == "" {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "grammar_markdown is required",
			Code:  "MISSING_GRAMMAR",
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

	if err := appConfig.Repository.UpdateLevel(number, req.GrammarMD); err != nil {
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
			Korean:   v.Korean,
			English:  v.English,
			Category: v.Category,
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

type extractFromURLRequest struct {
	URL string `json:"url"`
}

func ExtractFromUrlHandler(c *gin.Context) {
	var req extractFromURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid request body", Code: "INVALID_REQUEST"})
		return
	}
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "url is required", Code: "INVALID_URL"})
		return
	}

	parsed, err := url.Parse(req.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid URL scheme, must be http or https", Code: "INVALID_URL"})
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(req.URL)
	if err != nil {
		c.JSON(http.StatusBadGateway, errorResponse{Error: fmt.Sprintf("failed to fetch URL: %v", err), Code: "URL_FETCH_FAILED"})
		return
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, 100*1024)
	body, err := io.ReadAll(limited)
	if err != nil {
		c.JSON(http.StatusBadGateway, errorResponse{Error: "failed to read URL response", Code: "URL_FETCH_FAILED"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, errorResponse{
			Error: fmt.Sprintf("URL returned status %d", resp.StatusCode),
			Code:  "URL_FETCH_FAILED",
		})
		return
	}

	html := string(body)
	if strings.Contains(html, "too large") || len(body) >= 100*1024 {
		// Check if reader was truncated (technically not perfect but rare case)
	}

	result, err := sengen.ExtractFromHTML(html)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: fmt.Sprintf("extraction failed: %v", err),
			Code:  "EXTRACTION_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grammar_markdown": result.GrammarMD,
		"vocabulary":       result.Vocabulary,
	})
}

func GetCategoriesHandler(c *gin.Context) {
	categories, err := appConfig.Repository.Categories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to fetch categories",
			Code:  "CATEGORIES_ERROR",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
