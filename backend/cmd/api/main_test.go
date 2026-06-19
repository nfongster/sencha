package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type mockLLMServer struct {
	server *httptest.Server
	URL    string
}

func startMockLLM(t *testing.T) *mockLLMServer {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": "저는 학생입니다\nI am a student\n그녀는 의사입니다\nShe is a doctor",
					},
				},
			},
		})
	})
	server := httptest.NewServer(mux)
	return &mockLLMServer{server: server, URL: server.URL}
}

func (m *mockLLMServer) Close() {
	m.server.Close()
}

func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "server")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestMain_WrongDirectory_ExitsError(t *testing.T) {
	bin := buildBinary(t)
	workDir := t.TempDir()
	cmd := exec.Command(bin)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit, got success")
	}
	if !strings.Contains(string(out), "could not load config") {
		t.Fatalf("expected error about config load failure, got:\n%s", out)
	}
}

func TestMain_CorrectDirectory_StartsSuccessfully(t *testing.T) {
	bin := buildBinary(t)

	mock := startMockLLM(t)
	defer mock.Close()

	workDir := t.TempDir()
	configJSON := fmt.Sprintf(`{"llm": {"base_url": "%s/v1", "model": "test-model"}}`, mock.URL)
	if err := os.WriteFile(filepath.Join(workDir, "config.json"), []byte(configJSON), 0644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()

	if err != nil && !strings.Contains(string(out), "Starting server on :8080") {
		t.Fatalf("expected server to start, got error=%v output=\n%s", err, out)
	}
	if !strings.Contains(string(out), "Starting server on :8080") {
		t.Fatalf("expected 'Starting server on :8080' in output, got:\n%s", out)
	}
}
