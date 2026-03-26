package handlers

import (
	"net/http"

	"github.com/AlexanderVitto/padel-booking-api/internal/db/queries"
	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileHandler struct {
	pool *pgxpool.Pool
}

func NewProfileHandler(pool *pgxpool.Pool) *ProfileHandler {
	return &ProfileHandler{pool: pool}
}

// GetMe mengembalikan profile user yang sedang login.
func (h *ProfileHandler) GetMe(c *gin.Context) {
	userID, ok := mustGetUserID(c)
	if !ok {
		return
	}

	u, err := queries.GetUserByID(c.Request.Context(), h.pool, userID)
	if err != nil {
		if err == queries.ErrUserNotFound {
			// seharusnya tidak terjadi — user ada di JWT tapi tidak ada di DB
			// kemungkinan user sudah dihapus
			response.Error(c, http.StatusUnauthorized, "unauthorized", "user not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get profile")
		return
	}

	response.OK(c, gin.H{"user": u})
}

// updateMeRequest — semua field opsional.
// Minimal salah satu harus diisi (divalidasi di handler).
type updateMeRequest struct {
	// DisplayName opsional — kalau diisi, update display name
	DisplayName *string `json:"display_name"`
	// NewPassword opsional — kalau diisi, wajib sertakan CurrentPassword
	NewPassword *string `json:"new_password"`
	// CurrentPassword wajib kalau NewPassword diisi
	// untuk verifikasi identitas sebelum ganti password
	CurrentPassword *string `json:"current_password"`
}

// UpdateMe mengupdate profile user yang sedang login.
func (h *ProfileHandler) UpdateMe(c *gin.Context) {
	userID, ok := mustGetUserID(c)
	if !ok {
		return
	}

	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid request body")
		return
	}

	// minimal satu field harus diisi
	if req.DisplayName == nil && req.NewPassword == nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "at least one field must be provided: display_name or new_password")
		return
	}

	// kalau mau ganti password → wajib verifikasi password lama
	if req.NewPassword != nil {
		if req.CurrentPassword == nil {
			response.Error(c, http.StatusBadRequest, "validation_error", "current_password is required to change password")
			return
		}
		if len(*req.NewPassword) < 8 {
			response.Error(c, http.StatusBadRequest, "validation_error", "new_password must be at least 8 characters")
			return
		}

		// ambil user beserta password_hash untuk verifikasi
		u, err := queries.GetUserWithHashByID(c.Request.Context(), h.pool, userID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to verify current password")
			return
		}

		if err := queries.CheckPassword(u, *req.CurrentPassword); err != nil {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "current_password is incorrect")
			return
		}
	}

	// display_name validation kalau diisi
	if req.DisplayName != nil && len(*req.DisplayName) == 0 {
		response.Error(c, http.StatusBadRequest, "validation_error", "display_name cannot be empty")
		return
	}

	u, err := queries.UpdateUser(c.Request.Context(), h.pool, queries.UpdateUserParams{
		ID:          userID,
		DisplayName: req.DisplayName,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		if err == queries.ErrUserNotFound {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "user not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to update profile")
		return
	}

	response.OK(c, gin.H{"user": u})
}
