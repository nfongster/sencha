package sengen

import (
	"bytes"
	"context"
	"encoding/json"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"sencha/backend/internal/config"
	"sencha/backend/internal/session"
)

//go:embed grammar.md
var grammarMD string

//go:embed prompt.tmpl
var promptTmplSrc string

type promptData struct {
	Count   int
	Grammar string
	Vocab   string
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

var globalConfig *config.LLMConfig
var httpClient = &http.Client{Timeout: 60 * time.Second}

// testGenerateFunc, when set, replaces the real generation logic for tests.
// This allows handler tests to avoid depending on an actual LLM.
var testGenerateFunc func(int) ([]session.SentencePair, error)

func SetGenerateFunc(fn func(int) ([]session.SentencePair, error)) {
	testGenerateFunc = fn
}

func Init(cfg *config.LLMConfig) {
	globalConfig = cfg
}

func Generate(count int) ([]session.SentencePair, error) {
	if testGenerateFunc != nil {
		return testGenerateFunc(count)
	}
	if globalConfig == nil {
		return nil, fmt.Errorf("sengen not initialized")
	}

	prompt, err := buildPrompt(count)
	if err != nil {
		return nil, fmt.Errorf("building prompt: %w", err)
	}

	reply, err := callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	pairs, err := parseResponse(reply)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", err)
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("LLM returned no sentences")
	}

	return pairs, nil
}

func buildPrompt(count int) (string, error) {
	tmpl, err := template.New("prompt").Parse(promptTmplSrc)
	if err != nil {
		return "", fmt.Errorf("parsing prompt template: %w", err)
	}

	var vocabBuf strings.Builder
	for _, entry := range vocabList {
		vocabBuf.WriteString(fmt.Sprintf("- %s = %s\n", entry.Korean, entry.English))
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, promptData{
		Count:   count,
		Grammar: grammarMD,
		Vocab:   vocabBuf.String(),
	})
	if err != nil {
		return "", fmt.Errorf("executing prompt template: %w", err)
	}

	return buf.String(), nil
}

func callLLM(prompt string) (string, error) {
	body := chatRequest{
		Model: globalConfig.Model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are a Korean language teacher creating practice sentences for students."},
			{Role: "user", Content: prompt},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	url := strings.TrimRight(globalConfig.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if globalConfig.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+globalConfig.APIKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(raw, &chatResp); err != nil {
		return "", fmt.Errorf("parsing response JSON: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func parseResponse(text string) ([]session.SentencePair, error) {
	raw := strings.Split(strings.TrimSpace(text), "\n")

	var lines []string
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}

	var pairs []session.SentencePair
	for i := 0; i+1 < len(lines); i += 2 {
		korean := strings.Trim(lines[i], "\"「」『』")
		english := strings.Trim(lines[i+1], "\"「」『』")

		pairs = append(pairs, session.SentencePair{
			Korean:  korean,
			English: english,
		})
	}

	return pairs, nil
}
