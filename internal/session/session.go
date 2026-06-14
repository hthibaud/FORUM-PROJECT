package session

import (
	"Forum/internal/db"
	"Forum/pkg/utils"
	"errors"
	"net/http"
	"time"
)

const sessionDuration = 1 * time.Hour

// CreateSession creates a new session for a given user and stores it in the database
func CreateSession(w http.ResponseWriter, userID int, r *http.Request) error {
	sessionID, err := utils.GenerateUUID()
	if err != nil {
		return errors.New("Unable to generate the session ID")
	}

	token, err := utils.GenerateUUID() // Utiliser un UUID comme token
	if err != nil {
		return errors.New("unable to generate the session token")
	}

	ip := r.RemoteAddr

	err = db.CreateSession(sessionID, token, userID, sessionDuration, ip)
	if err != nil {
		utils.LogError("Unable to create the session as an DB file", err)
		return errors.New("internal server error")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(sessionDuration),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// GetUserIDFromSession récupère l'ID de l'utilisateur à partir du cookie de session en vérifiant la base de données
func GetUserIDFromSession(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, errors.New("cookie de session non trouvé")
	}
	sessionID := cookie.Value

	session, err := db.GetSessionByUUID(sessionID)
	if err != nil {
		utils.LogError("Erreur lors de la récupération de la session en bdd", err)
		return 0, errors.New("erreur interne du serveur")
	}

	if session == nil || time.Now().After(session.EndAt) {
		if session != nil {
			// Nettoyer la session expirée
			err := db.DeleteSessionByUUID(sessionID)
			if err != nil {
				utils.LogError("Impossible de supprimer la session expirée", err)
			}
		}
		return 0, errors.New("session invalide ou expirée")
	}

	return session.UserID, nil
}

// DeleteSession supprime la session de l'utilisateur de la base de données
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return
	}
	sessionID := cookie.Value

	err = db.DeleteSessionByUUID(sessionID)
	if err != nil {
		utils.LogError("Impossible de supprimer la session de la bdd", err)
	}

	// Expire le cookie dans le navigateur
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// IsAuthenticated vérifie si l'utilisateur est authentifié
func IsAuthenticated(r *http.Request) bool {
	_, err := GetUserIDFromSession(r)
	return err == nil
}
