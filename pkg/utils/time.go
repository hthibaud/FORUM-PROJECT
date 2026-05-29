package utils

import "fmt"

// MsToTime convertit une durée en millisecondes vers un format string "MM:SS".
func MsToTime(ms int) string {
	minutes := ms / 60000
	seconds := (ms % 60000) / 1000
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
