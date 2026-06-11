package config

import (
	"os"
)

func Init() {

	GitHub = GitHubConfig{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("GITHUB_REDIRECT_URI"),
	}
}
