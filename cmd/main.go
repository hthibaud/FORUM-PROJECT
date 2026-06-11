package main

import (
	"Forum/internal/config"
	"Forum/internal/router"
	"Forum/pkg/utils"
)

func main() {
	utils.SetDebugMode(true)

	config.Init()
	router.Start()
}
