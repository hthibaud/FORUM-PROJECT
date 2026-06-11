package session

import (
	"net/http"
	"sync"
	"time"
)

var (
	sessions      = make(map[string]string)
	sessionsMutex sync.RWMutex
)

func GetOrCreateSession(w http.ResponseWriter, r *http.Request) string {
	var sessionID string

	// Vérifie si le cookie de session existe
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionID = cookie.Value
	}

	sessionsMutex.RLock()
	token, exists := sessions[sessionID]
	sessionsMutex.RUnlock()

	// Création d'une nouvelle session si absente ou invalide
	if !exists {
		//sessionID = UUID
		//token = UUID

		// Sauvegarde sécurisée en mémoire (thread-safe)
		sessionsMutex.Lock()
		sessions[sessionID] = token
		sessionsMutex.Unlock()

		// Configuration du cookie de session
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			Expires:  time.Now().Add(1 * time.Hour),
			HttpOnly: true,  // Protège contre XSS
			Secure:   false, // HTTPS
			SameSite: http.SameSiteLaxMode,
		})
	}

	return token
}
