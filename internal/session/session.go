package session

import (
	"Forum/internal/db"
	"Forum/pkg/utils"
	"errors"
	"fmt"
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

	token, err := utils.GenerateUUID()
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

// GetUserIDFromSession retrieves the user’s ID from the session cookie by checking the database
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
			err := db.DeleteSessionByUUID(sessionID)
			if err != nil {
				utils.LogError("Impossible de supprimer la session expirée", err)
			}
		}
		return 0, errors.New("session invalide ou expirée")
	}

	return session.UserID, nil
}

// DeleteSession deletes the user’s session from the database
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

	// Expire the cookie in the browser
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(r *http.Request) bool {
	_, err := GetUserIDFromSession(r)
	return err == nil
}

// GetUserFromSession retrieves user information from the current session.
func GetUserFromSession(r *http.Request) (*db.User, error) {
	utils.Debug("GetUserFromSession: Attempting to get user from session")
	userID, err := GetUserIDFromSession(r)
	if err != nil {
		utils.Debug(fmt.Sprintf("GetUserFromSession: Could not get user ID from session: %v", err))
		return nil, nil // No valid session, so no user
	}

	utils.Debug(fmt.Sprintf("GetUserFromSession: Session valid for user ID: %d", userID))
	user, err := db.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("could not get user by id %d: %w", userID, err)
	}

	if user == nil {
		utils.Debug(fmt.Sprintf("GetUserFromSession: GetUserByID returned no user for ID: %d", userID))
		return nil, nil
	}

	utils.Debug(fmt.Sprintf("GetUserFromSession: Successfully retrieved user '%s'", user.Username))
	return user, nil
}

// StartSessionCleanup starts a cleanup routine for expired sessions.
func startSessionCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			utils.Log("Running session cleanup...")
			err := db.DeleteExpiredSessions()
			if err != nil {
				utils.LogError("Error during session cleanup", err)
			}
		}
	}()
}

func Init() {
	startSessionCleanup()
}
