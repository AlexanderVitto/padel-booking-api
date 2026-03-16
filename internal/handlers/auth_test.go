package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// helper untuk buat AuthHandler test tanpa DB
func newTestAuthHandler() *AuthHandler {
	return &AuthHandler{
		pool:             nil,
		jwtAccessSecret:  []byte("access-secret"),
		jwtRefreshSecret: []byte("refresh-secret"),
		jwtAccessTTL:     15 * time.Minute,
		jwtRefreshTTL:    30 * 24 * time.Hour,
	}
}

// ── Register ───────────────────────────────────────────────────────────────

func TestRegister_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	h := newTestAuthHandler()

	body, _ := json.Marshal(map[string]string{
		"email":        "not-an-email",
		"password":     "password123",
		"display_name": "John",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	h := newTestAuthHandler()

	body, _ := json.Marshal(map[string]string{
		"email":        "user@example.com",
		"password":     "short",
		"display_name": "John",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Login ─────────────────────────────────────────────────────────────────

func TestLogin_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_InvalidEmail(t *testing.T) {
	h := newTestAuthHandler()

	body, _ := json.Marshal(map[string]string{
		"email":    "not-an-email",
		"password": "password123",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Refresh ───────────────────────────────────────────────────────────────

func TestRefresh_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Refresh(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	h := newTestAuthHandler()

	body, _ := json.Marshal(map[string]string{
		"refresh_token": "this-is-not-a-valid-jwt",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Refresh(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ── Logout ────────────────────────────────────────────────────────────────

func TestLogout_InvalidBody(t *testing.T) {
	h := newTestAuthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/auth/logout", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Logout(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
