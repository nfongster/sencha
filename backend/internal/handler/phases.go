package handler

import (
	"net/http"
	"strconv"

	"sencha/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func ListPhasesHandler(c *gin.Context) {
	phases, err := appConfig.Repository.ListPhases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to list phases",
			Code:  "PHASES_LIST_ERROR",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"phases": phases})
}

type createPhaseRequest struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}

func CreatePhaseHandler(c *gin.Context) {
	var req createPhaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "phase name is required",
			Code:  "MISSING_NAME",
		})
		return
	}

	maxPhase, err := appConfig.Repository.MaxPhaseNumber()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "failed to check phase number",
			Code:  "PHASE_NUMBER_ERROR",
		})
		return
	}

	if req.Number < 1 || req.Number > maxPhase+1 {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "phase number must be between 1 and " + strconv.Itoa(maxPhase+1),
			Code:  "INVALID_PHASE_NUMBER",
		})
		return
	}

	if err := appConfig.Repository.CreatePhase(repository.Phase{
		Number: req.Number,
		Name:   req.Name,
	}); err != nil {
		c.JSON(http.StatusConflict, errorResponse{
			Error: err.Error(),
			Code:  "PHASE_CREATE_ERROR",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "phase created"})
}

func LevelsInPhaseHandler(c *gin.Context) {
	phaseNumberStr := c.Param("number")
	phaseNumber, err := strconv.Atoi(phaseNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			Error: "invalid phase number",
			Code:  "INVALID_PHASE_NUMBER",
		})
		return
	}

	levels, err := appConfig.Repository.LevelsInPhase(phaseNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			Error: "phase not found",
			Code:  "PHASE_NOT_FOUND",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"levels": levels})
}
