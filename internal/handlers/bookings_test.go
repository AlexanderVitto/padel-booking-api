package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setUserContext adalah helper untuk simulasi RequireJWT middleware di test.
// Set user_id ke gin context seperti yang dilakukan middleware aslinya.
func setUserContext(c *gin.Context, userID string) {
	c.Set("user_id", userID)
}

const testUserID = "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

// ── List ──────────────────────────────────────────────────────────────────

func TestList_MissingUserContext(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings", nil)

	// sengaja tidak set user_id di context → simulasi middleware tidak jalan
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestList_InvalidCourtID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings?court_id=not-a-uuid", nil)
	setUserContext(c, testUserID)

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestList_InvalidDate(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings?date=not-a-date", nil)
	setUserContext(c, testUserID)

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestList_InvalidLimit(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings?limit=999", nil)
	setUserContext(c, testUserID)

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestList_InvalidCursor(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings?cursor=not-a-timestamp", nil)
	setUserContext(c, testUserID)

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────

func TestGetByID_MissingUserContext(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings/some-id", nil)

	h.GetByID(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetByID_InvalidUUID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	setUserContext(c, testUserID)

	h.GetByID(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────────

func TestCreate_MissingUserContext(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/bookings", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestCreate_InvalidBody(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/bookings", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreate_InvalidCourtID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	body, _ := json.Marshal(map[string]string{
		"court_id":   "not-a-uuid",
		"start_time": "2026-12-01T10:00:00Z",
		"end_time":   "2026-12-01T11:00:00Z",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/bookings", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreate_EndTimeBeforeStartTime(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	body, _ := json.Marshal(map[string]string{
		"court_id":   "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"start_time": "2026-12-01T11:00:00Z",
		"end_time":   "2026-12-01T10:00:00Z", // end sebelum start
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/bookings", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreate_StartTimeInPast(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	body, _ := json.Marshal(map[string]string{
		"court_id":   "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"start_time": "2020-01-01T10:00:00Z", // di masa lalu
		"end_time":   "2020-01-01T11:00:00Z",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/bookings", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	setUserContext(c, testUserID)

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Cancel ────────────────────────────────────────────────────────────────

func TestCancel_MissingUserContext(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/bookings/some-id/cancel", nil)

	h.Cancel(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestCancel_InvalidUUID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/bookings/not-a-uuid/cancel", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}
	setUserContext(c, testUserID)

	h.Cancel(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
