package main

import (
	"Forum/internal/config"
	"Forum/internal/db"
	"Forum/internal/router"
	"Forum/internal/session"
	"Forum/pkg/openenv"
	"Forum/pkg/utils"
)

func main() {
	utils.SetDebugMode(false)

	config.Init()
	utils.SetDebugMode(config.Config.DEBUG)
	openenv.Init()
	utils.LoadTemplates()

	db.Init()
	defer db.Close()

	session.Init()
	router.Start()

}
