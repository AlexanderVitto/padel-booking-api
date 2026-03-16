package handlers

import (
	"github.com/google/uuid"
)

// isValidUUID cek apakah string adalah UUID v4 yang valid.
func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
