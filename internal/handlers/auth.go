package handlers

import (
	"net/http"
	"time"

	"github.com/AlexanderVitto/padel-booking-api/internal/config"
	"github.com/AlexanderVitto/padel-booking-api/internal/db/queries"
	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	pool             *pgxpool.Pool
	jwtAccessSecret  []byte
	jwtRefreshSecret []byte
	jwtAccessTTL     time.Duration
	jwtRefreshTTL    time.Duration
}

func NewAuthHandler(pool *pgxpool.Pool, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		pool:             pool,
		jwtAccessSecret:  []byte(cfg.JWTAccessSecret),
		jwtRefreshSecret: []byte(cfg.JWTRefreshSecret),
		jwtAccessTTL:     cfg.JWTAccessTTL,
		jwtRefreshTTL:    cfg.JWTRefreshTTL,
	}
}

// ── Register ───────────────────────────────────────────────────────────────

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := queries.CreateUser(c.Request.Context(), h.pool, queries.CreateUserParams{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		if err == queries.ErrEmailTaken {
			response.Error(c, http.StatusConflict, "email_taken", "Email is already registered")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to register user")
		return
	}

	accessToken, refreshToken, err := h.generateTokenPair(c, user)
	if err != nil {
		return // generateTokenPair sudah handle response error
	}

	response.Created(c, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ── Login ─────────────────────────────────────────────────────────────────

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, err := queries.GetUserByEmail(c.Request.Context(), h.pool, req.Email)
	if err != nil {
		if err == queries.ErrUserNotFound {
			// jangan expose "user not found" → pakai pesan generic untuk keamanan
			response.Error(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to login")
		return
	}

	if err := queries.CheckPassword(user, req.Password); err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		return
	}

	accessToken, refreshToken, err := h.generateTokenPair(c, user)
	if err != nil {
		return // generateTokenPair sudah handle response error
	}

	response.OK(c, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ── Refresh ───────────────────────────────────────────────────────────────

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// 1. validasi JWT refresh token
	claims, err := h.parseRefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid_token", "invalid or expired refresh token")
		return
	}

	userID := claims["sub"].(string)

	// 2. cek token ada di DB (belum di-revoke & belum expired)
	_, err = queries.VerifyRefreshToken(c.Request.Context(), h.pool, req.RefreshToken)
	if err != nil {
		if err == queries.ErrRefreshTokenNotFound || err == queries.ErrRefreshTokenExpired {
			response.Error(c, http.StatusUnauthorized, "invalid_token", "invalid or expired refresh token")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to verify token")
		return
	}

	// 3. ambil data user terbaru dari DB
	user, err := queries.GetUserByID(c.Request.Context(), h.pool, userID)
	if err != nil {
		if err == queries.ErrUserNotFound {
			response.Error(c, http.StatusUnauthorized, "invalid_token", "user not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get user")
		return
	}

	// 4. buat access token baru
	accessToken, err := h.generateAccessToken(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to generate token")
		return
	}

	response.OK(c, gin.H{
		"access_token": accessToken,
	})
}

// ── Logout ────────────────────────────────────────────────────────────────

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// hapus refresh token dari DB (revoke)
	// tidak perlu error kalau token tidak ditemukan — idempotent
	_ = queries.DeleteRefreshToken(c.Request.Context(), h.pool, req.RefreshToken)

	c.Status(http.StatusNoContent) // 204 — berhasil, tidak ada body
}

// ── JWT Helper ────────────────────────────────────────────────────────

type jwtClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

// generateTokenPair membuat access token + refresh token sekaligus,
// menyimpan refresh token ke DB, dan return keduanya.
func (h *AuthHandler) generateTokenPair(c *gin.Context, user queries.User) (accessToken, refreshToken string, err error) {
	accessToken, err = h.generateAccessToken(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to generate access token")
		return
	}

	refreshToken, err = h.generateRefreshToken(user)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to generate refresh token")
		return
	}

	// simpan refresh token ke DB
	expiresAt := time.Now().Add(h.jwtRefreshTTL)
	if err = queries.SaveRefreshToken(c.Request.Context(), h.pool, user.ID, refreshToken, expiresAt); err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to save refresh token")
		return
	}

	return accessToken, refreshToken, nil
}

func (h *AuthHandler) generateAccessToken(user queries.User) (string, error) {
	claims := jwtClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.jwtAccessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(h.jwtAccessSecret)

}

func (h *AuthHandler) generateRefreshToken(user queries.User) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.jwtRefreshTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(h.jwtRefreshSecret)
}

func (h *AuthHandler) parseRefreshToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return h.jwtRefreshSecret, nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil

}
