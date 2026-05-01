package middleware

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"bank-sampah-backend/internal/repository"
	"bank-sampah-backend/pkg/crypto"
	"bank-sampah-backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

// HMACAuth validates HMAC-SHA256 signatures from SIMAK
func HMACAuth(schoolRepo *repository.SchoolRepository, callbackRepo *repository.CallbackRepository, toleranceSec int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract required headers
		apiKey := c.Get("X-School-Key")
		signature := c.Get("X-Signature")
		timestamp := c.Get("X-Timestamp")
		nonce := c.Get("X-Nonce")

		if apiKey == "" || signature == "" || timestamp == "" || nonce == "" {
			return response.Unauthorized(c, "Header keamanan tidak lengkap (X-School-Key, X-Signature, X-Timestamp, X-Nonce)")
		}

		// 1. Validate timestamp — prevent replay attacks
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return response.BadRequest(c, "Format timestamp tidak valid")
		}

		diff := math.Abs(float64(time.Now().Unix() - ts))
		if diff > float64(toleranceSec) {
			return response.Unauthorized(c, fmt.Sprintf("Request expired. Toleransi: %d detik", toleranceSec))
		}

		// 2. Check nonce — prevent replay
		nonceUsed, err := callbackRepo.NonceExists(nonce)
		if err != nil {
			return response.InternalError(c, "Gagal memeriksa nonce")
		}
		if nonceUsed {
			return response.Conflict(c, "Nonce sudah pernah digunakan (replay detected)")
		}

		// 3. Lookup school by API key
		school, err := schoolRepo.FindByAPIKey(apiKey)
		if err != nil {
			return response.Unauthorized(c, "API Key tidak valid atau sekolah tidak aktif")
		}

		// 4. Reconstruct canonical string and verify HMAC
		bodyHash := crypto.HashSHA256(string(c.Body()))
		canonical := crypto.BuildCanonicalString(
			c.Method(),
			c.Path(),
			timestamp,
			nonce,
			bodyHash,
		)

		fmt.Printf("DEBUG GO HMAC:\nMethod: %s\nPath: %s\nTS: %s\nNonce: %s\nBodyHash: %s\nSignature: %s\nExpected: %s\nCanonical:\n%s\n---\n", 
			c.Method(), c.Path(), timestamp, nonce, bodyHash, signature, crypto.GenerateHMAC(school.APISecret, canonical), canonical)

		if !crypto.VerifyHMAC(school.APISecret, canonical, signature) {
			return response.Unauthorized(c, "Signature HMAC tidak valid")
		}

		// 5. Store nonce to prevent replay (TTL = tolerance * 2)
		ttl := time.Duration(toleranceSec*2) * time.Second
		if err := callbackRepo.SaveNonce(nonce, ttl); err != nil {
			// Log but don't fail — nonce collision is non-critical
			fmt.Printf("⚠️ Gagal menyimpan nonce: %v\n", err)
		}

		// 6. Store school info in context
		c.Locals("school_id", school.ID.String())
		c.Locals("school_name", school.Name)
		c.Locals("school", school)

		return c.Next()
	}
}
