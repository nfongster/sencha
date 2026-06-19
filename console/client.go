package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultServerURL = "http://localhost:8080"

const (
	routeSessions = "/api/sessions"
	routeReveal   = "/api/sessions/%s/reveal"
	routeGrade    = "/api/sessions/%s/grade"
	routeLevel    = "/api/levels/%d"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultServerURL
	}
	return &Client{
		baseURL: baseURL,
		http:    http.DefaultClient,
	}
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

type gradeResponse struct {
	CardsRemaining  int            `json:"cards_remaining"`
	SessionComplete bool           `json:"session_complete"`
	GradeCounts     map[string]int `json:"grade_counts"`
}

type apiError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func (c *Client) CreateSession(direction string) (*sessionResponse, error) {
	body := bytes.NewBufferString(`{"direction":"` + direction + `"}`)
	resp, err := c.http.Post(c.baseURL+routeSessions, "application/json", body)
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

func (c *Client) RevealCard(sessionID string) (*revealResponse, error) {
	url := c.baseURL + fmt.Sprintf(routeReveal, sessionID)
	resp, err := c.http.Post(url, "application/json", nil)
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

type level struct {
	Number       int    `json:"number"`
	PhaseNumber  int    `json:"phase_number"`
	GrammarMD    string `json:"grammar_md"`
	ExceptionsMD string `json:"exceptions_md"`
}

type ruleUpdateRequest struct {
	GrammarMD  string `json:"grammar_markdown"`
	Exceptions string `json:"exceptions_markdown"`
}

func (c *Client) GetLevel(number int) (*level, error) {
	resp, err := c.http.Get(c.baseURL + fmt.Sprintf(routeLevel, number))
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}

	var payload struct {
		Level level `json:"level"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &payload.Level, nil
}

func (c *Client) UpdateLevel(number int, grammarMD, exceptions string) error {
	body, _ := json.Marshal(ruleUpdateRequest{
		GrammarMD:  grammarMD,
		Exceptions: exceptions,
	})
	url := c.baseURL + fmt.Sprintf(routeLevel, number)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}
	return nil
}

func (c *Client) SubmitGrade(sessionID, grade string) (*gradeResponse, error) {
	url := c.baseURL + fmt.Sprintf(routeGrade, sessionID)
	body := bytes.NewBufferString(`{"grade":"` + grade + `"}`)
	resp, err := c.http.Post(url, "application/json", body)
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
