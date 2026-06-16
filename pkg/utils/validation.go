package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// TrimWhitespace removes leading and trailing whitespace from a string.
func TrimWhitespace(value string) string {
	return strings.TrimSpace(value)
}

// NormalizeWhitespace trims the string and replaces multiple spaces with a single space.
func NormalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// ValidateNotEmpty checks that a field value is not empty.
func ValidateNotEmpty(value, fieldName string) string {
	if TrimWhitespace(value) == "" {
		return fmt.Sprintf("The %s field is required.", fieldName)
	}
	return ""
}

// ValidateLength checks minimum and maximum length constraints.
func ValidateLength(value, fieldName string, min, max int) string {
	length := len(value)
	if min > 0 && length < min {
		return fmt.Sprintf("%s must be at least %d characters long.", fieldName, min)
	}
	if max > 0 && length > max {
		return fmt.Sprintf("%s cannot exceed %d characters.", fieldName, max)
	}
	return ""
}

// ValidateEmail checks that an email address is present and has a simple valid format.
func ValidateEmail(value, fieldName string) string {
	if err := ValidateNotEmpty(value, fieldName); err != "" {
		return err
	}
	if err := ValidateLength(value, fieldName, 5, 254); err != "" {
		return err
	}
	if !emailPattern.MatchString(value) {
		return fmt.Sprintf("%s must be a valid email address.", fieldName)
	}
	return ""
}
