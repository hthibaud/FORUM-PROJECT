package router

import (
	"Forum/internal/config"
	"Forum/pkg/utils"
	"fmt"
	"log"
	"net/http"
)

func Start() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", home)
    mux.HandleFunc("/demo", demo)
    mux.HandleFunc("/forbidden", forbidden)
    mux.HandleFunc("/server-error", serverError)
    mux.HandleFunc("/register", register)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	utils.Log(fmt.Sprintf("Server started at : http://localhost:%v", config.Config.PORT))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.PORT), mux))
}
