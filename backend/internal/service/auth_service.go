package service

import (
	"errors"
	"time"

	"bank-sampah-backend/internal/config"
	"bank-sampah-backend/internal/model"
	"bank-sampah-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	adminRepo *repository.AdminRepository
	jwtCfg    config.JWTConfig
}

func NewAuthService(adminRepo *repository.AdminRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		adminRepo: adminRepo,
		jwtCfg:    jwtCfg,
	}
}

// TokenPair holds access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// JWTClaims extends standard JWT claims with admin info
type JWTClaims struct {
	AdminID  string `json:"admin_id"`
	Username string `json:"username"`
	Type     string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Login authenticates an admin and returns JWT tokens
func (s *AuthService) Login(username, password string) (*TokenPair, error) {
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("username atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("username atau password salah")
	}

	return s.generateTokenPair(admin)
}

// RefreshToken validates a refresh token and returns a new access token
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	if claims.Type != "refresh" {
		return nil, errors.New("token bukan refresh token")
	}

	adminID, err := uuid.Parse(claims.AdminID)
	if err != nil {
		return nil, errors.New("admin ID tidak valid")
	}

	admin, err := s.adminRepo.FindByID(adminID)
	if err != nil {
		return nil, errors.New("admin tidak ditemukan")
	}

	return s.generateTokenPair(admin)
}

// ValidateToken parses and validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak valid")
		}
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}

// SeedAdmin creates the initial admin account if none exists
func (s *AuthService) SeedAdmin(username, email, password string) error {
	exists, err := s.adminRepo.ExistsAny()
	if err != nil {
		return err
	}
	if exists {
		return nil // Already seeded
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.Admin{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
	}

	return s.adminRepo.Create(admin)
}

func (s *AuthService) generateTokenPair(admin *model.Admin) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.jwtCfg.AccessExpiry)

	// Access Token
	accessClaims := &JWTClaims{
		AdminID:  admin.ID.String(),
		Username: admin.Username,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "bank-sampah",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, err
	}

	// Refresh Token
	refreshClaims := &JWTClaims{
		AdminID:  admin.ID.String(),
		Username: admin.Username,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtCfg.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "bank-sampah",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}
