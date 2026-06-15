package utils

import "github.com/google/uuid"

// GenerateUUID generates a new UUID string.
func GenerateUUID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
