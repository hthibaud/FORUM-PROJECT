package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// TrimWhitespace nettoie les espaces inutiles en début et fin de chaîne.
func TrimWhitespace(value string) string {
	return strings.TrimSpace(value)
}

// NormalizeWhitespace nettoie la chaîne et remplace les espaces multiples par un seul espace.
func NormalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// ValidateNotEmpty vérifie qu'une valeur de champ n'est pas vide.
func ValidateNotEmpty(value, fieldName string) string {
	if TrimWhitespace(value) == "" {
		return fmt.Sprintf("Le champ %s est obligatoire.", fieldName)
	}
	return ""
}

// ValidateLength vérifie les contraintes minimales et maximales de longueur.
func ValidateLength(value, fieldName string, min, max int) string {
	length := len(value)
	if min > 0 && length < min {
		return fmt.Sprintf("%s doit contenir au moins %d caractères.", fieldName, min)
	}
	if max > 0 && length > max {
		return fmt.Sprintf("%s ne peut pas dépasser %d caractères.", fieldName, max)
	}
	return ""
}

// ValidateEmail vérifie qu'une adresse e-mail est présente et a un format simple valide.
func ValidateEmail(value, fieldName string) string {
	if err := ValidateNotEmpty(value, fieldName); err != "" {
		return err
	}
	if err := ValidateLength(value, fieldName, 5, 254); err != "" {
		return err
	}
	if !emailPattern.MatchString(value) {
		return fmt.Sprintf("%s doit être une adresse e-mail valide.", fieldName)
	}
	return ""
}
