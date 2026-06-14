package main

import (
	"Forum/internal/config"
	"Forum/internal/db"
	"Forum/internal/router"
	"Forum/pkg/utils"
)

func main() {
	utils.SetDebugMode(false)

	config.Init()
	utils.SetDebugMode(config.Config.DEBUG)
	utils.LoadTemplates()
	db.Init()
	defer db.Close()
	router.Start()

}
