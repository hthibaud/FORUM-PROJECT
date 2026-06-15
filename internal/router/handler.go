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
	mux.HandleFunc("/not-found", notFound)
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", logout)
	mux.HandleFunc("/post/", postView)
	mux.HandleFunc("/post/create", createPost)
	mux.HandleFunc("/category/", categoryPage)
	// Like / Dislike handlers
	mux.HandleFunc("/like/post", handlePostLike)
	mux.HandleFunc("/like/comment", handleCommentLike)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	utils.Log(fmt.Sprintf("Server started at : http://localhost:%v", config.Config.PORT))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.PORT), mux))
}
