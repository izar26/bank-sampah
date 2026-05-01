package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError holds field-level validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(msgs, "; ")
}

// IsEmpty checks if there are no validation errors
func (ve ValidationErrors) IsEmpty() bool {
	return len(ve) == 0
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateRequired checks if a string field is non-empty
func ValidateRequired(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Field: field, Message: "field ini wajib diisi"}
	}
	return nil
}

// ValidateEmail checks if a string is a valid email format
func ValidateEmail(field, value string) *ValidationError {
	if !emailRegex.MatchString(value) {
		return &ValidationError{Field: field, Message: "format email tidak valid"}
	}
	return nil
}

// ValidateMinLength checks minimum string length
func ValidateMinLength(field, value string, min int) *ValidationError {
	if len(value) < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("minimal %d karakter", min),
		}
	}
	return nil
}

// ValidatePositive checks if amount is positive
func ValidatePositive(field string, value int64) *ValidationError {
	if value <= 0 {
		return &ValidationError{Field: field, Message: "harus bernilai positif"}
	}
	return nil
}
