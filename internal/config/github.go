package config

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

var GitHub GitHubConfig
