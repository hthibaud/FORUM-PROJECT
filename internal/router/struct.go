package router

import "Forum/internal/db"

// PageData holds the data to be passed to the HTML templates.
type PageData struct {
	Title            string
	Message          string
	User             db.User
	Categories       []db.Category
	SelectedCategory *db.Category
	Posts            []db.Post
	Notifications    []db.Notification
	IsAuthenticated  bool
	Post             db.Post
	Comments         []*db.Comment
	Reports          []*db.Report
}

type discoverPageData struct {
	Title      string
	Categories []db.Category
}
