package handler

import (
	"bank-sampah-backend/internal/service"
	"bank-sampah-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Login handles admin authentication
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Format request tidak valid")
	}

	if req.Username == "" || req.Password == "" {
		return response.BadRequest(c, "Username dan password wajib diisi")
	}

	tokens, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.Success(c, "Login berhasil", tokens)
}

// RefreshToken generates a new access token from a refresh token
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Format request tidak valid")
	}

	if req.RefreshToken == "" {
		return response.BadRequest(c, "Refresh token wajib diisi")
	}

	tokens, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.Success(c, "Token berhasil diperbarui", tokens)
}

// Me returns the current authenticated admin info
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	adminID := c.Locals("admin_id")
	username := c.Locals("admin_username")

	return response.Success(c, "Data admin", fiber.Map{
		"admin_id": adminID,
		"username": username,
	})
}
