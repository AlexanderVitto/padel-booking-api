package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", true},
		{"22222222-2222-2222-2222-222222222222", true},
		{"not-a-uuid", false},
		{"", false},
		{"12345678-1234-1234-1234-12345678901z", false},
	}

	for _, tt := range tests {
		got := isValidUUID(tt.input)
		if got != tt.want {
			t.Errorf("isValidUUID(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParseIntQuery(t *testing.T) {
	tests := []struct {
		input    string
		fallback int
		want     int
	}{
		{"10", 20, 10},
		{"", 20, 20},
		{"abc", 20, 20},
		{"0", 20, 0},
		{"-5", 20, -5},
	}

	for _, tt := range tests {
		got := parseIntQuery(tt.input, tt.fallback)
		if got != tt.want {
			t.Errorf("parseIntQuery(%q, %d) = %d, want %d", tt.input, tt.fallback, got, tt.want)
		}
	}
}

func TestList_MissingCourtID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings?date=2026-02-28", nil)

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestList_InvalidCourtIDUUID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings?court_id=not-a-uuid&date=2026-02-28", nil)
	c.Params = gin.Params{{Key: "court_id", Value: "not-a-uuid"}}

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestList_MissingDate(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings?court_id=22222222-2222-2222-2222-222222222222", nil)

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetByID_InvalidUUID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/bookings/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.GetByID(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCancel_InvalidUUID(t *testing.T) {
	h := &BookingsHandler{pool: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v1/bookings/not-a-uuid/cancel", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Cancel(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
