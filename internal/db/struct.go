package db

import (
	"database/sql"
	"time"
)

// Category represents a single category from the database.
type Category struct {
	ID   int
	Name string
	Desc string
}

// User represents a single user from the database.
type User struct {
	ID        int
	Username  string
	Password  string // This will be the hashed password
	Email     string
	CreatedAt time.Time
}

// Session represents a user session in the database.
type Session struct {
	ID        int
	UUID      string
	Token     string
	UserID    int
	CreatedAt time.Time
	EndAt     time.Time
	IP        string
}

// Post represents a single forum post.
type Post struct {
	ID          int
	AuthorID    int
	CategoryID  int
	Title       string
	Text        string
	DataUUID    string
	Timestamp   time.Time
	Author      string // Username of the author
	CommentsNbr int    // Number of comments on the post
}

// Comment represents a single comment on a post.
type Comment struct {
	ID             int
	UserID         int
	PostID         int
	RepID          sql.NullInt64 // Can be NULL
	Text           string
	Timestamp      time.Time
	AuthorUsername string // Username of the author
}
