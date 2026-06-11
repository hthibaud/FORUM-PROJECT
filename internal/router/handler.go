package router

import (
	"Forum/pkg/utils"
	"log"
	"net/http"
)

func Start() {
	http.HandleFunc("/", home)
	http.HandleFunc("/demo", demo)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	utils.Log("Server started at : http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
