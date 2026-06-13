package db

import (
	"Forum/pkg/utils"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const dbPath = "./data/database.db"

// db is the package-level database connection pool.
var db *sql.DB

// Init initializes the database connection, creates tables if they don't exist,
// and loads initial data. It should be called once at application startup.
func Init() {
	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	isNewDB := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		isNewDB = true
		utils.Log("The database file does not exist. It will be created.")
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if isNewDB {
		utils.Debug("Initializing a Blank Database...")
	} else {
		utils.Debug("The file already exists. Checking the integrity of the tables...")
	}

	for tableName, createQuery := range schema {
		exists, err := tableExists(tableName)
		if err != nil {
			log.Fatalf("Error when checking the table %s: %v", tableName, err)
		}

		if !exists {
			utils.Debug(fmt.Sprintf("The table '%s' is missing. Creating...\n", tableName))
			if _, err := db.Exec(createQuery); err != nil {
				log.Fatalf("Error creating the table %s: %v", tableName, err)
			}
			utils.Debug(fmt.Sprintf("Table '%s' created successfully.\n", tableName))
		} else {
			utils.Debug(fmt.Sprintf("Table '%s' already present. RAS.\n", tableName))
		}
	}

	if !loadCategories() {
		log.Fatal("Error loading and initializing categories")
	}

	utils.Log("[OK] Database ready and verified.")
}

// Close closes the database connection. It should be deferred in main().
func Close() {
	if db != nil {
		db.Close()
	}
}

// tableExists checks if a table exists in the database.
func tableExists(tableName string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?;`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// loadCategories loads initial data into the 'cat' table.
func loadCategories() bool {
	for name, desc := range categories {
		var exists bool
		err := db.QueryRow("SELECT 1 FROM cat WHERE name = ?", name).Scan(&exists)

		if err != nil && err != sql.ErrNoRows {
			utils.LogError(fmt.Sprintf("Error checking if category '%s' exists: %v", name), err)
			return false
		}

		if err == sql.ErrNoRows {
			if _, err := db.Exec("INSERT INTO cat (name, desc) VALUES (?, ?)", name, desc); err != nil {
				utils.LogError(fmt.Sprintf("Error inserting category '%s': %v", name), err)
				return false
			} else {
				utils.Debug(fmt.Sprintf("Category '%s' inserted.", name))
			}
		}
	}
	return true
}

// GetCategories retrieves all categories from the database.
func GetCategories() ([]Category, error) {
	rows, err := db.Query("SELECT id, name, desc FROM cat")
	if err != nil {
		return nil, fmt.Errorf("could not query categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Desc); err != nil {
			return nil, fmt.Errorf("could not scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return categories, nil
}

// -- User Functions --

// CreateUser adds a new user to the database with a hashed password.
func CreateUser(username, password, email string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("could not hash password: %w", err)
	}

	query := `INSERT INTO users (username, password, email) VALUES (?, ?, ?)`
	_, err = db.Exec(query, username, hashedPassword, email)
	if err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}
	return nil
}

// GetUserByUsername retrieves a user by their username.
func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, created_at FROM users WHERE username = ?`
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if no user is found
		}
		return nil, fmt.Errorf("could not get user by username: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, created_at FROM users WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get user by id: %w", err)
	}
	return user, nil
}

// -- Session Functions --

// CreateSession creates a new session for a user.
func CreateSession(uuid, token string, userID int, duration time.Duration, ip string) error {
	endAt := time.Now().Add(duration)
	query := `INSERT INTO session (uuid, token, user_id, created_at, end_at, ip) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, uuid, token, userID, time.Now(), endAt, ip)
	if err != nil {
		return fmt.Errorf("could not create session: %w", err)
	}
	return nil
}

// GetSessionByUUID retrieves a session by its UUID.
func GetSessionByUUID(uuid string) (*Session, error) {
	session := &Session{}
	query := `SELECT id, uuid, token, user_id, created_at, end_at, ip FROM session WHERE uuid = ?`
	err := db.QueryRow(query, uuid).Scan(&session.ID, &session.UUID, &session.Token, &session.UserID, &session.CreatedAt, &session.EndAt, &session.IP)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get session by uuid: %w", err)
	}
	return session, nil
}

// DeleteSessionByUUID deletes a session from the database.
func DeleteSessionByUUID(uuid string) error {
	query := `DELETE FROM session WHERE uuid = ?`
	_, err := db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("could not delete session: %w", err)
	}
	return nil
}

// -- Post Functions --

// CreatePost adds a new post to the database.
func CreatePost(authorID, categoryID int, title, text string) error {
	dataUUID, err := utils.GenerateUUID() // Assuming you have a UUID generator
	if err != nil {
		return fmt.Errorf("could not generate UUID for post: %w", err)
	}

	query := `INSERT INTO post (author, category_id, title, text, data_uuid) VALUES (?, ?, ?, ?, ?)`
	_, err = db.Exec(query, authorID, categoryID, title, text, dataUUID)
	if err != nil {
		return fmt.Errorf("could not create post: %w", err)
	}
	return nil
}

// GetPostByID retrieves a single post by its ID, including the author's username.
func GetPostByID(id int) (*Post, error) {
	post := &Post{}
	query := `
		SELECT p.id, p.author, p.category_id, p.title, p.text, p.data_uuid, p.timestamp, u.username
		FROM post p
		JOIN users u ON p.author = u.id
		WHERE p.id = ?`
	err := db.QueryRow(query, id).Scan(&post.ID, &post.AuthorID, &post.CategoryID, &post.Title, &post.Text, &post.DataUUID, &post.Timestamp, &post.Author)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get post by id: %w", err)
	}
	return post, nil
}

// GetPostsByCategory retrieves all posts for a given category.
// Deprecated: Use GetPostsByCategoryPaginated for better performance.
func GetPostsByCategory(categoryID int) ([]Post, error) {
	query := `
		SELECT p.id, p.author, p.category_id, p.title, p.timestamp, u.username
		FROM post p
		JOIN users u ON p.author = u.id
		WHERE p.category_id = ?
		ORDER BY p.timestamp DESC`

	rows, err := db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("could not query posts by category: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.CategoryID, &post.Title, &post.Timestamp, &post.Author); err != nil {
			return nil, fmt.Errorf("could not scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

// GetRecentPosts retrieves a limited number of the most recent posts across all categories.
func GetRecentPosts(limit int) ([]Post, error) {
	query := `
		SELECT p.id, p.title, p.timestamp, u.username
		FROM post p
		JOIN users u ON p.author = u.id
		ORDER BY p.timestamp DESC
		LIMIT ?`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("could not query recent posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Timestamp, &post.Author); err != nil {
			return nil, fmt.Errorf("could not scan recent post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

// GetPostsByCategoryPaginated retrieves a slice of posts for a given category with a limit and offset.
func GetPostsByCategoryPaginated(categoryID, limit, offset int) ([]Post, error) {
	query := `
		SELECT p.id, p.author, p.category_id, p.title, p.timestamp, u.username
		FROM post p
		JOIN users u ON p.author = u.id
		WHERE p.category_id = ?
		ORDER BY p.timestamp DESC
		LIMIT ? OFFSET ?`

	rows, err := db.Query(query, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("could not query posts by category: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.CategoryID, &post.Title, &post.Timestamp, &post.Author); err != nil {
			return nil, fmt.Errorf("could not scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

// -- Comment Functions --

// CreateComment adds a new comment to a post.
func CreateComment(userID, postID int, repID sql.NullInt64, text string) error {
	query := `INSERT INTO post_message (user_id, post_id, rep_id, text) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, userID, postID, repID, text)
	if err != nil {
		return fmt.Errorf("could not create comment: %w", err)
	}
	return nil
}

// GetCommentsByPostID retrieves all comments for a given post.
func GetCommentsByPostID(postID int) ([]Comment, error) {
	query := `
		SELECT c.id, c.user_id, c.post_id, c.rep_id, c.text, c.timestamp, u.username
		FROM post_message c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.timestamp ASC`

	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("could not query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.RepID, &comment.Text, &comment.Timestamp, &comment.AuthorUsername); err != nil {
			return nil, fmt.Errorf("could not scan comment: %w", err)
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

// GetTopLevelCommentsByPostID retrieves paginated top-level comments (not replies) for a post.
func GetTopLevelCommentsByPostID(postID, limit, offset int) ([]Comment, error) {
	query := `
		SELECT c.id, c.user_id, c.post_id, c.rep_id, c.text, c.timestamp, u.username
		FROM post_message c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ? AND c.rep_id IS NULL
		ORDER BY c.timestamp ASC
		LIMIT ? OFFSET ?`

	rows, err := db.Query(query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("could not query top-level comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.RepID, &comment.Text, &comment.Timestamp, &comment.AuthorUsername); err != nil {
			return nil, fmt.Errorf("could not scan comment: %w", err)
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

// GetChildComments retrieves all direct replies to a parent comment.
func GetChildComments(parentID int) ([]Comment, error) {
	query := `
		SELECT c.id, c.user_id, c.post_id, c.rep_id, c.text, c.timestamp, u.username
		FROM post_message c
		JOIN users u ON c.user_id = u.id
		WHERE c.rep_id = ?
		ORDER BY c.timestamp ASC`

	rows, err := db.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("could not query child comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.RepID, &comment.Text, &comment.Timestamp, &comment.AuthorUsername); err != nil {
			return nil, fmt.Errorf("could not scan child comment: %w", err)
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

// GetParentComment retrieves the parent of a given comment.
// Returns nil, nil if the comment is a top-level comment.
func GetParentComment(childID int) (*Comment, error) {
	var parentID sql.NullInt64
	err := db.QueryRow("SELECT rep_id FROM post_message WHERE id = ?", childID).Scan(&parentID)
	if err != nil {
		return nil, fmt.Errorf("could not find child comment: %w", err)
	}

	if !parentID.Valid {
		return nil, nil // This is a top-level comment, no parent.
	}

	parentComment := &Comment{}
	query := `
		SELECT c.id, c.user_id, c.post_id, c.rep_id, c.text, c.timestamp, u.username
		FROM post_message c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?`
	err = db.QueryRow(query, parentID.Int64).Scan(&parentComment.ID, &parentComment.UserID, &parentComment.PostID, &parentComment.RepID, &parentComment.Text, &parentComment.Timestamp, &parentComment.AuthorUsername)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve parent comment: %w", err)
	}

	return parentComment, nil
}
