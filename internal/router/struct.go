package router

import "Forum/internal/db"

// PageData holds the data to be passed to the HTML templates.
type PageData struct {
	Title           string
	Message         string
	Categories      []db.Category
	IsAuthenticated bool
}

type discoverPageData struct {
	Title      string
	Categories []db.Category
}
