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
	Role      string
	IsBanned  bool
	CreatedAt time.Time
}

// Report represents a report from a user.
type Report struct {
	ID              int
	ReporterID      int
	ContentID       int
	ContentType     string
	Reason          string
	CreatedAt       time.Time
	ReporterName    string
	Status          string
	ContentAuthorID int
	ParentPostID    int // ID du post parent (pour les commentaires)
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
	Likes       int
	Dislikes    int
	UserChoice  int // 1 for like, -1 for dislike, 0 for no vote
}

// Comment represents a single comment on a post.
type Comment struct {
	ID             int
	AuthorID       int
	AuthorUsername string
	PostID         int
	ParentID       sql.NullInt64 // Use sql.NullInt64 for nullable foreign keys
	Text           string
	Timestamp      time.Time
	Replies        []*Comment // For nested comments
	Likes          int
	Dislikes       int
	UserChoice     int // 1 for like, -1 for dislike, 0 for no vote
}

// Notification represents a user notification.
type Notification struct {
	ID        int
	UserID    int
	ActorID   int
	ActorName string
	PostID    sql.NullInt64
	Message   string
	IsRead    bool
	CreatedAt time.Time
}
