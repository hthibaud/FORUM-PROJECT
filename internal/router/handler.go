package router

import (
	"Forum/internal/config"
	"Forum/pkg/utils"
	"fmt"
	"log"
	"net/http"
)

func Start() {
	http.HandleFunc("/", home)
	http.HandleFunc("/demo", demo)
	http.HandleFunc("/register", register)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	utils.Log(fmt.Sprintf("Server started at : http://localhost:%v", config.Config.PORT))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.PORT), nil))
}
