package sengen

import (
	"bytes"
	"context"
	"encoding/json"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"sencha/backend/internal/config"
	"sencha/backend/internal/repository"
	"sencha/backend/internal/session"
)

//go:embed prompt.tmpl
var embeddedPromptTmpl string

//go:embed gramchecker.tmpl
var gramCheckTmplSrc string

var promptTmplSrc string

type promptData struct {
	Count   int
	Grammar string
	Vocab   string
}

type gramCheckData struct {
	Pairs string
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

func init() {
	loadPrompt()
}

// testGenerateFunc, when set, replaces the real generation logic for tests.
// This allows handler tests to avoid depending on an actual LLM.
type GenerateFunc func(int, repository.LevelData) ([]session.SentencePair, error)

var testGenerateFunc GenerateFunc

func SetGenerateFunc(fn GenerateFunc) {
	testGenerateFunc = fn
}

func loadPrompt() {
	data, err := os.ReadFile("prompt.tmpl")
	if err != nil {
		promptTmplSrc = embeddedPromptTmpl
		return
	}
	promptTmplSrc = string(data)
}

func ReloadPrompt() error {
	data, err := os.ReadFile("prompt.tmpl")
	if err != nil {
		return fmt.Errorf("reading prompt template on disk: %w", err)
	}
	promptTmplSrc = string(data)
	return nil
}

func GetPrompt() string {
	return promptTmplSrc
}

func Init(cfg *config.LLMConfig) {
	log.Printf("[sengen] Init called — base_url=%q model=%q api_key=%q", cfg.BaseURL, cfg.Model, cfg.APIKey)
	globalConfig = cfg
	loadPrompt()
}

func Generate(count int, data repository.LevelData) ([]session.SentencePair, error) {
	if testGenerateFunc != nil {
		return testGenerateFunc(count, data)
	}
	if globalConfig == nil {
		return nil, fmt.Errorf("sengen not initialized")
	}

	log.Printf("[sengen] Generate(%d) starting — using base_url=%q model=%q", count, globalConfig.BaseURL, globalConfig.Model)

	prompt, err := buildPrompt(count, data)
	if err != nil {
		return nil, fmt.Errorf("building prompt: %w", err)
	}

	genStart := time.Now()
	reply, err := callLLM(prompt)
	genLatency := time.Since(genStart)
	if err != nil {
		log.Printf("[sengen] Generate: callLLM failed: %v", err)
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	log.Printf("[sengen] Generation step complete — latency=%.3fms", float64(genLatency.Microseconds())/1000)

	pairs, err := parseResponse(reply)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", err)
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("LLM returned no sentences")
	}

	checked, gramLatency, err := grammarCheck(pairs)
	if err != nil {
		log.Printf("[sengen] Grammar check failed: %v", err)
		return nil, fmt.Errorf("grammar check failed: %w", err)
	}

	log.Printf("[sengen] Grammar check step complete — latency=%.3fms", float64(gramLatency.Microseconds())/1000)

	return checked, nil
}

func buildPrompt(count int, data repository.LevelData) (string, error) {
	tmpl, err := template.New("prompt").Parse(promptTmplSrc)
	if err != nil {
		return "", fmt.Errorf("parsing prompt template: %w", err)
	}

	var vocabBuf strings.Builder
	for _, entry := range data.Vocab {
		vocabBuf.WriteString(fmt.Sprintf("- %s = %s\n", entry.Korean, entry.English))
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, promptData{
		Count:   count,
		Grammar: data.GrammarMD,
		Vocab:   vocabBuf.String(),
	})
	if err != nil {
		return "", fmt.Errorf("executing prompt template: %w", err)
	}

	return buf.String(), nil
}

func buildGrammarCheckPrompt(pairs []session.SentencePair) (string, error) {
	tmpl, err := template.New("gramcheck").Parse(gramCheckTmplSrc)
	if err != nil {
		return "", fmt.Errorf("parsing grammar check template: %w", err)
	}

	var lines []string
	for _, p := range pairs {
		lines = append(lines, fmt.Sprintf("%q\n%q", p.Korean, p.English))
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, gramCheckData{
		Pairs: strings.Join(lines, "\n"),
	})
	if err != nil {
		return "", fmt.Errorf("executing grammar check template: %w", err)
	}

	return buf.String(), nil
}

func grammarCheck(pairs []session.SentencePair) ([]session.SentencePair, time.Duration, error) {
	prompt, err := buildGrammarCheckPrompt(pairs)
	if err != nil {
		return nil, 0, fmt.Errorf("building grammar check prompt: %w", err)
	}

	start := time.Now()
	reply, err := callLLM(prompt)
	latency := time.Since(start)
	if err != nil {
		return nil, 0, fmt.Errorf("grammar check LLM call failed: %w", err)
	}

	checked, err := parseResponse(reply)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing grammar check response: %w", err)
	}

	if len(checked) == 0 {
		return nil, 0, fmt.Errorf("grammar check returned no sentences")
	}

	return checked, latency, nil
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
	log.Printf("[sengen] callLLM: POST %s | model=%q | body size=%d bytes", url, globalConfig.Model, len(payload))

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
		log.Printf("[sengen] callLLM: httpClient.Do error: %v", err)
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[sengen] callLLM: io.ReadAll error: %v", err)
		return "", fmt.Errorf("reading response: %w", err)
	}

	log.Printf("[sengen] callLLM: response status=%d body=%s", resp.StatusCode, truncate(string(raw), 200))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(raw, &chatResp); err != nil {
		log.Printf("[sengen] callLLM: json.Unmarshal failed: %v | raw=%s", err, truncate(string(raw), 500))
		return "", fmt.Errorf("parsing response JSON: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// ── URL Extraction ──

//go:embed extract.tmpl
var extractTmplSrc string

type ExtractResult struct {
	GrammarMD  string                   `json:"grammar_markdown"`
	Vocabulary []repository.VocabEntry  `json:"vocabulary"`
}

type extractData struct {
	HTML string
}

type ExtractFunc func(string) (*ExtractResult, error)

var testExtractFunc ExtractFunc

func SetExtractFunc(fn ExtractFunc) {
	testExtractFunc = fn
}

func ExtractFromHTML(html string) (*ExtractResult, error) {
	if testExtractFunc != nil {
		return testExtractFunc(html)
	}
	if globalConfig == nil {
		return nil, fmt.Errorf("sengen not initialized")
	}

	tmpl, err := template.New("extract").Parse(extractTmplSrc)
	if err != nil {
		return nil, fmt.Errorf("parsing extract template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, extractData{HTML: html}); err != nil {
		return nil, fmt.Errorf("executing extract template: %w", err)
	}

	reply, err := callLLM(buf.String())
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	var result ExtractResult
	if err := json.Unmarshal([]byte(reply), &result); err != nil {
		return nil, fmt.Errorf("parsing LLM response as JSON: %w", err)
	}

	if result.GrammarMD == "" {
		return nil, fmt.Errorf("LLM returned no grammar rules")
	}

	return &result, nil
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
