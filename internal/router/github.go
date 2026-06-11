package router

import (
	"Forum/internal/config"
	"fmt"
	"net/http"
)

func GitHubLogin(w http.ResponseWriter, r *http.Request) {

	url := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s",
		config.GitHub.ClientID,
	)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
