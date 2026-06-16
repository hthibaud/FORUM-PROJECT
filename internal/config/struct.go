package config

type configData struct {
	DEBUG      bool `json:"DEBUG"`
	HTTP_PORT  uint `json:"HTTP"`
	HTTPS_PORT uint `json:"HTTPS"`
}
