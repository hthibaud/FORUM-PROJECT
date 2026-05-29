package main

import (
	"Forum/internal/config"
	"Forum/internal/router"
)

func main() {
	config.Init()
	router.Start()
}
