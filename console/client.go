package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const serverURL = "http://localhost:8080"

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

type gradeResponse struct {
	CardsRemaining  int            `json:"cards_remaining"`
	SessionComplete bool           `json:"session_complete"`
	GradeCounts     map[string]int `json:"grade_counts"`
}

type apiError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func createSession(direction string) (*sessionResponse, error) {
	body := bytes.NewBufferString(`{"direction":"` + direction + `"}`)
	resp, err := http.Post(serverURL+"/api/sessions", "application/json", body)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}

	var sess sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &sess, nil
}

func revealCard(sessionID string) (*revealResponse, error) {
	resp, err := http.Post(serverURL+"/api/sessions/"+sessionID+"/reveal", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}

	var reveal revealResponse
	if err := json.NewDecoder(resp.Body).Decode(&reveal); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &reveal, nil
}

func submitGrade(sessionID, grade string) (*gradeResponse, error) {
	body := bytes.NewBufferString(`{"grade":"` + grade + `"}`)
	resp, err := http.Post(serverURL+"/api/sessions/"+sessionID+"/grade", "application/json", body)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}

	var result gradeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &result, nil
}
