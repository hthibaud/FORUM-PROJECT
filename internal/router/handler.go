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
	mux.HandleFunc("/forbidden", forbidden)
	mux.HandleFunc("/server-error", serverError)
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
	mux.HandleFunc("/general", general)
	mux.HandleFunc("/post/", postView)
	mux.HandleFunc("/post/create", createPost)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	utils.Log(fmt.Sprintf("Server started at : http://localhost:%v", config.Config.PORT))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.PORT), mux))
}
