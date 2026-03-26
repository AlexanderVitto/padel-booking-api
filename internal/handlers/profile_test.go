package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// ── GetMe ─────────────────────────────────────────────────────────────────

func TestGetMe_MissingUserContext(t *testing.T) {
	h := &ProfileHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/me", nil)

	// sengaja tidak set user_id → simulasi middleware tidak jalan
	h.GetMe(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── UpdateMe ──────────────────────────────────────────────────────────────

func TestUpdateMe_MissingUserContext(t *testing.T) {
	h := &ProfileHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/me", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateMe(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestUpdateMe_NoFieldProvided(t *testing.T) {
	h := &ProfileHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/me", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.UpdateMe(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_NewPasswordWithoutCurrentPassword(t *testing.T) {
	h := &ProfileHandler{pool: nil}

	body, _ := json.Marshal(map[string]string{
		"new_password": "newpassword123",
		// sengaja tidak kirim current_password
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.UpdateMe(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_NewPasswordTooShort(t *testing.T) {
	h := &ProfileHandler{pool: nil}

	body, _ := json.Marshal(map[string]string{
		"new_password":     "short", // kurang dari 8 karakter
		"current_password": "oldpassword123",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.UpdateMe(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateMe_EmptyDisplayName(t *testing.T) {
	h := &ProfileHandler{pool: nil}

	emptyName := ""
	body, _ := json.Marshal(map[string]*string{
		"display_name": &emptyName,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/me", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.UpdateMe(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
