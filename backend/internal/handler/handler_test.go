package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"sencha/backend/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	store.Reset()
	RegisterRoutes(r)
	return r
}

func TestHealthEndpoint(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}

func TestCreateSession_DefaultDirection(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sessions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp sessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.SessionID)
	assert.Equal(t, "korean-to-english", resp.Direction)
	assert.Equal(t, 10, resp.TotalCards)
	assert.Equal(t, 10, resp.CardsRemaining)
	assert.False(t, resp.SessionComplete)
}

func TestCreateSession_WithDirection(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"direction":"english-to-korean"}`)
	req, _ := http.NewRequest("POST", "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp sessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "english-to-korean", resp.Direction)
}

func TestCreateSession_InvalidDirection(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"direction":"invalid"}`)
	req, _ := http.NewRequest("POST", "/api/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSession_NotFound(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/sessions/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetSession_ReturnsStatus(t *testing.T) {
	r := setupRouter()
	createW := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sessions", nil)
	r.ServeHTTP(createW, req)

	var created sessionResponse
	json.Unmarshal(createW.Body.Bytes(), &created)

	w := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/sessions/"+created.SessionID, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp sessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, created.SessionID, resp.SessionID)
	assert.Equal(t, 10, resp.CardsRemaining)
}

func TestReveal_ReturnsCard(t *testing.T) {
	r := setupRouter()
	sessionID := createSession(t, r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/reveal", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp revealResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotZero(t, resp.CardID)
	assert.NotEmpty(t, resp.Front)
	assert.NotEmpty(t, resp.Back)
}

func TestReveal_NotFound(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sessions/nonexistent/reveal", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGrade_AfterReveal(t *testing.T) {
	r := setupRouter()
	sessionID := createSession(t, r)

	revealW := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/reveal", nil)
	r.ServeHTTP(revealW, req)
	assert.Equal(t, http.StatusOK, revealW.Code)

	gradeW := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"grade":"pass"}`)
	req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/grade", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(gradeW, req)

	assert.Equal(t, http.StatusOK, gradeW.Code)

	var resp gradeResponse
	err := json.Unmarshal(gradeW.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 9, resp.CardsRemaining)
	assert.False(t, resp.SessionComplete)
	assert.Equal(t, 1, resp.GradeCounts["pass"])
}

func TestGrade_WithoutReveal_ReturnsConflict(t *testing.T) {
	r := setupRouter()
	sessionID := createSession(t, r)

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"grade":"pass"}`)
	req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/grade", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestGrade_InvalidGrade_ReturnsBadRequest(t *testing.T) {
	r := setupRouter()
	sessionID := createSession(t, r)

	revealW := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/reveal", nil)
	r.ServeHTTP(revealW, req)

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"grade":"invalid"}`)
	req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/grade", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGrade_NotFound(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"grade":"pass"}`)
	req, _ := http.NewRequest("POST", "/api/sessions/nonexistent/grade", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCompleteSession_ReturnsSummary(t *testing.T) {
	r := setupRouter()
	sessionID := createSession(t, r)

	for i := 0; i < 10; i++ {
		revealW := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/reveal", nil)
		r.ServeHTTP(revealW, req)
		assert.Equal(t, http.StatusOK, revealW.Code)

		gradeW := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"grade":"pass"}`)
		req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/grade", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(gradeW, req)
		assert.Equal(t, http.StatusOK, gradeW.Code)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/sessions/"+sessionID, nil)
	r.ServeHTTP(w, req)

	var resp sessionResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.SessionComplete)
	assert.Equal(t, 0, resp.CardsRemaining)
}

func TestActionsOnCompleteSession_ReturnsConflict(t *testing.T) {
	r := setupRouter()
	sessionID := createSession(t, r)

	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/reveal", nil)
		r.ServeHTTP(httptest.NewRecorder(), req)
		body := bytes.NewBufferString(`{"grade":"pass"}`)
		req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/grade", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	t.Run("reveal returns 409", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/reveal", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("grade returns 409", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"grade":"pass"}`)
		req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/grade", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestFullSessionFlow(t *testing.T) {
	r := setupRouter()
	sessionID := createSession(t, r)

	seen := make(map[int]bool)
	grades := []string{"pass", "hard", "fail", "pass", "hard", "fail", "pass", "pass", "hard", "fail"}
	for i, g := range grades {
		revealW := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/sessions/"+sessionID+"/reveal", nil)
		r.ServeHTTP(revealW, req)
		assert.Equal(t, http.StatusOK, revealW.Code, "reveal %d failed", i)

		var revealResp revealResponse
		json.Unmarshal(revealW.Body.Bytes(), &revealResp)
		assert.NotZero(t, revealResp.CardID)
		assert.False(t, seen[revealResp.CardID], "duplicate card %d at reveal %d", revealResp.CardID, i)
		seen[revealResp.CardID] = true
		assert.NotEmpty(t, revealResp.Front)
		assert.NotEmpty(t, revealResp.Back)

		gradeW := httptest.NewRecorder()
		body := bytes.NewBufferString(fmt.Sprintf(`{"grade":"%s"}`, g))
		req, _ = http.NewRequest("POST", "/api/sessions/"+sessionID+"/grade", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(gradeW, req)
		assert.Equal(t, http.StatusOK, gradeW.Code, "grade %d failed", i)
	}

	assert.Equal(t, 10, len(seen))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/sessions/"+sessionID, nil)
	r.ServeHTTP(w, req)

	var resp sessionResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.SessionComplete)
}

func createSession(t *testing.T, r *gin.Engine) string {
	t.Helper()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/sessions", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp sessionResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.SessionID
}
