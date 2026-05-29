package router

import (
	"Forum/pkg/utils"
	"log"
	"net/http"
)

func Start() {
	http.HandleFunc("/", home)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	utils.Log("Server started at : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
