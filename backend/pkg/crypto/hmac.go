package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// GenerateHMAC creates an HMAC-SHA256 signature for the given message
func GenerateHMAC(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMAC performs constant-time comparison of HMAC signatures
// to prevent timing attacks
func VerifyHMAC(secret, message, providedSignature string) bool {
	expectedSignature := GenerateHMAC(secret, message)
	return subtle.ConstantTimeCompare(
		[]byte(expectedSignature),
		[]byte(providedSignature),
	) == 1
}

// HashSHA256 returns the hex-encoded SHA256 hash of the input
func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// BuildCanonicalString creates the canonical string for HMAC signing
// Format: METHOD\nPATH\nTIMESTAMP\nNONCE\nBODY_HASH
func BuildCanonicalString(method, path, timestamp, nonce, bodyHash string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", method, path, timestamp, nonce, bodyHash)
}
