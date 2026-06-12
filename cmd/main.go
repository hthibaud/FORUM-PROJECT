package main

import (
	"Forum/internal/config"
	"Forum/internal/router"
	"Forum/pkg/utils"
	_ "github.com/mattn/go-sqlite3"

)

func main() {
	utils.SetDebugMode(false)

	config.Init()
	utils.SetDebugMode(config.Config.DEBUG)
	router.Start()

}
