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

	// Public routes that do not require ban checks for unauthenticated users,
	// but the middleware will handle redirection for banned users if they are logged in.
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/register", register)

	// Logout is handled inside the middleware to allow banned users to log out.
	mux.HandleFunc("/logout", logout)

	// Banned page itself should not have the middleware.
	mux.HandleFunc("/banned", bannedPage)

	// These handlers will be wrapped.
	mux.Handle("/", checkBannedStatus(http.HandlerFunc(home)))
	mux.Handle("/forbidden", checkBannedStatus(http.HandlerFunc(forbidden)))
	mux.Handle("/server-error", checkBannedStatus(http.HandlerFunc(serverError)))
	mux.Handle("/not-found", checkBannedStatus(http.HandlerFunc(notFound)))
	mux.Handle("/post/", checkBannedStatus(http.HandlerFunc(postView)))
	mux.Handle("/post/create", checkBannedStatus(http.HandlerFunc(createPost)))
	mux.Handle("/category/", checkBannedStatus(http.HandlerFunc(categoryPage)))
	mux.Handle("/like/post", checkBannedStatus(http.HandlerFunc(handlePostLike)))
	mux.Handle("/like/comment", checkBannedStatus(http.HandlerFunc(handleCommentLike)))
	mux.Handle("/report", checkBannedStatus(http.HandlerFunc(reportContent)))

	// Moderation handlers are wrapped in both middlewares.
	mux.Handle("/moderation", checkBannedStatus(isModerator(http.HandlerFunc(moderationPage))))
	mux.Handle("/moderation/ban/", checkBannedStatus(isModerator(http.HandlerFunc(banUser))))
	mux.Handle("/moderation/unban/", checkBannedStatus(isModerator(http.HandlerFunc(unbanUser))))
	mux.Handle("/moderation/handle/delete/", checkBannedStatus(isModerator(http.HandlerFunc(handleDeleteReport))))
	mux.Handle("/moderation/handle/dismiss/", checkBannedStatus(isModerator(http.HandlerFunc(handleDismissReport))))
	mux.Handle("/moderation/delete/post/", checkBannedStatus(isModerator(http.HandlerFunc(deletePost))))
	mux.Handle("/moderation/delete/comment/", checkBannedStatus(isModerator(http.HandlerFunc(deleteComment))))

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	utils.Log(fmt.Sprintf("Server started at : http://localhost:%v", config.Config.PORT))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.Config.PORT), mux))
}
