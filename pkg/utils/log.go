package utils

import (
	"fmt"
	"os"
	"time"
)

var (
	debugEnabled bool
)

// SetDebugMode active ou désactive l'affichage des logs de débogage.
func SetDebugMode(debug bool) {
	debugEnabled = debug
}

// Log affiche un message d'information standard dans la console avec horodatage.
func Log(info string) {
	date := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%v : %v\n", date, info)
}

// LogError affiche un message d'erreur en rouge dans la console avec horodatage.
func LogError(info string, err any) {
	date := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\033[31m %v : %v : %v\n\033[0m", date, info, err)
}

// LogFatal affiche un message d'erreur et arrête le programme.
func LogFatal(info string, err any) {
	date := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("\033[31mFATAL: %v : %v : %v\n\033[0m", date, info, err)
	os.Exit(1)
}

// Debug affiche un message de débogage si le mode debug est activé.
func Debug(info string) {
	if debugEnabled {
		date := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("(DEBUG) %v : %v\n", date, info)
	}

}
