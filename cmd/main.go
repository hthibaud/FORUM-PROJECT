package main

import (
	"Forum/internal/config"
	"Forum/internal/data"
	"Forum/internal/router"
	"Forum/pkg/utils"
	_ "github.com/mattn/go-sqlite3"

)

func main() {
	utils.SetDebugMode(true)

	config.Init()
	router.Start()
	
	db, err := data.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

}
