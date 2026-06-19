package handler

import (
	"fmt"
	"log"
	"net/http"

	"sencha/backend/internal/repository"
	"sencha/backend/internal/sengen"
	"sencha/backend/internal/session"
	"sencha/backend/internal/store"

	"github.com/gin-gonic/gin"
)

type createSessionRequest struct {
	Direction   string `json:"direction"`
	LevelNumber int    `json:"level_number"`
}

type sessionResponse struct {
	SessionID       string `json:"session_id"`
	Direction       string `json:"direction"`
	TotalCards      int    `json:"total_cards"`
	CardsRemaining  int    `json:"cards_remaining"`
	SessionComplete bool   `json:"session_complete"`
}

type revealResponse struct {
	CardID int    `json:"card_id"`
	Front  string `json:"front"`
	Back   string `json:"back"`
}

type gradeRequest struct {
	Grade string `json:"grade"`
}

type gradeResponse struct {
	CardsRemaining  int            `json:"cards_remaining"`
	SessionComplete bool           `json:"session_complete"`
	GradeCounts     map[string]int `json:"grade_counts"`
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func CreateSessionHandler(c *gin.Context) {
	var req createSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Direction = ""
	}

	dir := session.Direction(req.Direction)
	if req.Direction != "" && dir != session.DirectionKoreanToEnglish &&
		dir != session.DirectionEnglishToKorean && dir != session.DirectionMixed {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid direction, must be one of: korean-to-english, english-to-korean, mixed",
			Code:  "INVALID_DIRECTION",
		})
		return
	}

	levelNum := req.LevelNumber
	if levelNum < 1 {
		levelNum = 1
	}

	levelData, err := appConfig.Repository.LoadLevelData(levelNum)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: fmt.Sprintf("invalid level %d: %v", levelNum, err),
			Code:  "INVALID_LEVEL",
		})
		return
	}

	pairs, err := sengen.Generate(10, *levelData)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse{
			Error: "sentence generation failed",
			Code:  "GENERATION_FAILED",
		})
		return
	}

	var sess *session.Session
	if req.Direction != "" {
		sess = session.NewSession(session.WithDirection(dir), session.WithPairs(pairs))
	} else {
		sess = session.NewSession(session.WithPairs(pairs))
	}

	store.Set(sess.ID, sess)

	if err := appConfig.Repository.SaveSentences(sessionsToSentences(pairs, levelNum)); err != nil {
		log.Printf("[handler] failed to save sentences: %v", err)
	}

	c.JSON(http.StatusCreated, sessionResponse{
		SessionID:       sess.ID,
		Direction:       string(sess.Direction),
		TotalCards:      sess.TotalCards,
		CardsRemaining:  sess.CardsRemaining,
		SessionComplete: sess.SessionComplete,
	})
}

func sessionsToSentences(pairs []session.SentencePair, levelNum int) []repository.Sentence {
	sentences := make([]repository.Sentence, len(pairs))
	for i, p := range pairs {
		sentences[i] = repository.Sentence{
			LevelNumber: levelNum,
			Korean:      p.Korean,
			English:     p.English,
		}
	}
	return sentences
}

func GetSessionHandler(c *gin.Context) {
	id := c.Param("id")
	sess, ok := store.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "session not found",
			Code:  "NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, sessionResponse{
		SessionID:       sess.ID,
		Direction:       string(sess.Direction),
		TotalCards:      sess.TotalCards,
		CardsRemaining:  sess.CardsRemaining,
		SessionComplete: sess.SessionComplete,
	})
}

func RevealHandler(c *gin.Context) {
	id := c.Param("id")
	sess, ok := store.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "session not found",
			Code:  "NOT_FOUND",
		})
		return
	}

	card, err := sess.Reveal()
	if err != nil {
		if err == session.ErrSessionComplete {
			c.JSON(http.StatusConflict, errorResponse{
				Error: "session already complete",
				Code:  "SESSION_COMPLETE",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
			Code:  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, revealResponse{
		CardID: card.ID,
		Front:  card.Front,
		Back:   card.Back,
	})
}

func GradeHandler(c *gin.Context) {
	id := c.Param("id")
	sess, ok := store.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "session not found",
			Code:  "NOT_FOUND",
		})
		return
	}

	var req gradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	summary, err := sess.Grade(session.Grade(req.Grade))
	if err != nil {
		switch err {
		case session.ErrSessionComplete:
			c.JSON(http.StatusConflict, errorResponse{
				Error: "session already complete",
				Code:  "SESSION_COMPLETE",
			})
		case session.ErrCardNotRevealed:
			c.JSON(http.StatusConflict, errorResponse{
				Error: "card not yet revealed",
				Code:  "CARD_NOT_REVEALED",
			})
		case session.ErrInvalidGrade:
			c.JSON(http.StatusBadRequest, errorResponse{
				Error: "invalid grade, must be one of: pass, hard, fail",
				Code:  "INVALID_GRADE",
			})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
				Code:  "INTERNAL_ERROR",
			})
		}
		return
	}

	counts := make(map[string]int)
	for k, v := range summary.GradeCounts {
		counts[string(k)] = v
	}

	c.JSON(http.StatusOK, gradeResponse{
		CardsRemaining:  summary.CardsRemaining,
		SessionComplete: summary.SessionComplete,
		GradeCounts:     counts,
	})
}
