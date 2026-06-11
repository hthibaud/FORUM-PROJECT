package config

import (
	"Forum/pkg/utils"
	"encoding/json"
	"os"
)

var Config configData

func Init() {
	utils.Debug("Start config Init...")

	file, err := os.Open("config.json")
	if err != nil {
		utils.LogError("Error opening config file", err)
		os.Exit(1)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		utils.LogError("Config decode error", err)
		os.Exit(1)
	}

	utils.Debug("Config init succesfully !")
}
