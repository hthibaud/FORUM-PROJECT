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

	// Apply migrations for missing columns
	if err := applyMigrations(); err != nil {
		log.Fatalf("Error applying migrations: %v", err)
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

// columnExists checks if a column exists in a table.
func columnExists(tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var type_ string
		var notnull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &type_, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}
	return false, nil
}

// applyMigrations applies schema migrations to add missing columns.
func applyMigrations() error {
	// Check if the status column exists in the reports table
	exists, err := columnExists("reports", "status")
	if err != nil {
		return fmt.Errorf("could not check if status column exists: %w", err)
	}

	if !exists {
		utils.Debug("Adding missing 'status' column to 'reports' table...")
		if _, err := db.Exec("ALTER TABLE reports ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'"); err != nil {
			return fmt.Errorf("could not add status column to reports table: %w", err)
		}
		utils.Debug("Successfully added 'status' column to 'reports' table.")
	}

	return nil
}

// loadCategories loads initial data into the 'cat' table.
func loadCategories() bool {
	for name, desc := range categories {
		var exists bool
		err := db.QueryRow("SELECT 1 FROM cat WHERE name = ?", name).Scan(&exists)

		if err != nil && err != sql.ErrNoRows {
			utils.LogError(fmt.Sprintf("Error checking if category '%s' exists: %v", name, err), err)
			return false
		}

		if err == sql.ErrNoRows {
			if _, err := db.Exec("INSERT INTO cat (name, desc) VALUES (?, ?)", name, desc); err != nil {
				utils.LogError(fmt.Sprintf("Error inserting category '%s': %v", name, err), err)
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

// GetCategoryByID retrieves a single category by its ID.
func GetCategoryByID(id int) (*Category, error) {
	query := `SELECT id, name, desc FROM cat WHERE id = ?`
	row := db.QueryRow(query, id)

	category := &Category{}
	if err := row.Scan(&category.ID, &category.Name, &category.Desc); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get category by id: %w", err)
	}
	return category, nil
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
	query := `SELECT id, username, password, email, role, is_banned, created_at FROM users WHERE username = ?`
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.IsBanned, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if no user is found
		}
		return nil, fmt.Errorf("could not get user by username: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email address.
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, role, is_banned, created_at FROM users WHERE email = ?`
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.IsBanned, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get user by email: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID.
func GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, role, is_banned, created_at FROM users WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.IsBanned, &user.CreatedAt)
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

// DeleteExpiredSessions removes all expired sessions from the database.
func DeleteExpiredSessions() error {
	query := `DELETE FROM session WHERE end_at <= ?`
	_, err := db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("could not delete expired sessions: %w", err)
	}
	utils.Log("Expired sessions cleaned up.")
	return nil
}

// -- Post Functions --

// CreatePost adds a new post to the database.
func CreatePost(authorID, categoryID int, title, text string) (int64, error) {
	dataUUID, err := utils.GenerateUUID() // Assuming you have a UUID generator
	if err != nil {
		return 0, fmt.Errorf("could not generate UUID for post: %w", err)
	}

	query := `INSERT INTO post (author, category_id, title, text, data_uuid) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, authorID, categoryID, title, text, dataUUID)
	if err != nil {
		return 0, fmt.Errorf("could not create post: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("could not retrieve last insert ID: %w", err)
	}
	return id, nil
}

// GetPostByID retrieves a single post by its ID, including the author's username and like/dislike counts.
// The currentUserID is used to determine if the user has liked or disliked the post. Use 0 if the user is not logged in.
func GetPostByID(id int, currentUserID int) (*Post, error) {
	post := &Post{}
	query := `
		SELECT
			p.id, p.author, p.category_id, p.title, p.text, p.data_uuid, p.timestamp, u.username,
			COALESCE(SUM(CASE WHEN pl.type = 1 THEN 1 ELSE 0 END), 0) AS likes,
			COALESCE(SUM(CASE WHEN pl.type = -1 THEN 1 ELSE 0 END), 0) AS dislikes,
			COALESCE((SELECT type FROM post_likes WHERE user_id = ? AND post_id = p.id), 0) AS user_choice
		FROM post p
		JOIN users u ON p.author = u.id
		LEFT JOIN post_likes pl ON p.id = pl.post_id
		WHERE p.id = ?
		GROUP BY p.id`

	err := db.QueryRow(query, currentUserID, id).Scan(
		&post.ID, &post.AuthorID, &post.CategoryID, &post.Title, &post.Text,
		&post.DataUUID, &post.Timestamp, &post.Author, &post.Likes, &post.Dislikes, &post.UserChoice,
	)
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
		SELECT
			p.id, p.author, p.category_id, p.title, p.timestamp, u.username,
			(SELECT COUNT(*) FROM post_likes WHERE post_id = p.id AND type = 1) as likes,
			(SELECT COUNT(*) FROM post_likes WHERE post_id = p.id AND type = -1) as dislikes
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
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.CategoryID, &post.Title, &post.Timestamp, &post.Author, &post.Likes, &post.Dislikes); err != nil {
			return nil, fmt.Errorf("could not scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

// -- Comment Functions --

// CreateComment adds a new comment to a post.
func CreateComment(comment Comment) error {
	query := `INSERT INTO post_message (user_id, post_id, rep_id, text, timestamp) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, comment.AuthorID, comment.PostID, comment.ParentID, comment.Text, comment.Timestamp)
	if err != nil {
		return fmt.Errorf("could not create comment: %w", err)
	}

	// Trigger Notification
	if comment.ParentID.Valid {
		// Notify the parent comment author
		var parentAuthorID int
		err = db.QueryRow("SELECT user_id FROM post_message WHERE id = ?", comment.ParentID.Int64).Scan(&parentAuthorID)
		if err == nil {
			CreateNotification(parentAuthorID, comment.AuthorID, sql.NullInt64{Int64: int64(comment.PostID), Valid: true}, "a répondu à votre commentaire")
		}
	} else {
		// Notify the post author
		var postAuthorID int
		err = db.QueryRow("SELECT author FROM post WHERE id = ?", comment.PostID).Scan(&postAuthorID)
		if err == nil {
			CreateNotification(postAuthorID, comment.AuthorID, sql.NullInt64{Int64: int64(comment.PostID), Valid: true}, "a commenté votre post")
		}
	}

	return nil
}

// GetCommentsByPostID retrieves all comments for a given post, ordered by popularity.
func GetCommentsForPost(postID int, currentUserID int) ([]Comment, error) {
	query := `
		SELECT
			c.id, c.user_id, c.post_id, c.rep_id, c.text, c.timestamp, u.username,
			COALESCE(SUM(CASE WHEN cl.type = 1 THEN 1 ELSE 0 END), 0) AS likes,
			COALESCE(SUM(CASE WHEN cl.type = -1 THEN 1 ELSE 0 END), 0) AS dislikes,
			COALESCE((SELECT type FROM comment_likes WHERE user_id = ? AND comment_id = c.id), 0) AS user_choice
		FROM post_message c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN comment_likes cl ON c.id = cl.comment_id
		WHERE c.post_id = ?
		GROUP BY c.id
		ORDER BY (likes - dislikes) DESC, c.timestamp ASC`

	rows, err := db.Query(query, currentUserID, postID)
	if err != nil {
		return nil, fmt.Errorf("could not query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(
			&comment.ID, &comment.AuthorID, &comment.PostID, &comment.ParentID,
			&comment.Text, &comment.Timestamp, &comment.AuthorUsername,
			&comment.Likes, &comment.Dislikes, &comment.UserChoice,
		); err != nil {
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
		if err := rows.Scan(&comment.ID, &comment.AuthorID, &comment.PostID, &comment.ParentID, &comment.Text, &comment.Timestamp, &comment.AuthorUsername); err != nil {
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
		if err := rows.Scan(&comment.ID, &comment.AuthorID, &comment.PostID, &comment.ParentID, &comment.Text, &comment.Timestamp, &comment.AuthorUsername); err != nil {
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
	err = db.QueryRow(query, parentID.Int64).Scan(&parentComment.ID, &parentComment.AuthorID, &parentComment.PostID, &parentComment.ParentID, &parentComment.Text, &parentComment.Timestamp, &parentComment.AuthorUsername)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve parent comment: %w", err)
	}

	return parentComment, nil
}

// -- Like / Dislike Functions --

// LikePost applies a like or dislike to a post from a user.
func LikePost(userID, postID, likeType int) error {
	// The type should be 1 for a like, -1 for a dislike.
	// We first delete any existing like/dislike from this user for this post to avoid conflicts.
	// Then, we insert the new one. If the user clicks the same button again, the front-end should send a '0' to remove the vote.

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	// Remove any existing vote from this user for this post
	_, err = tx.Exec(`DELETE FROM post_likes WHERE user_id = ? AND post_id = ?`, userID, postID)
	if err != nil {
		return fmt.Errorf("could not remove existing post like: %w", err)
	}

	// If likeType is not 0, insert the new vote
	if likeType == 1 || likeType == -1 {
		_, err = tx.Exec(`INSERT INTO post_likes (user_id, post_id, type) VALUES (?, ?, ?)`, userID, postID, likeType)
		if err != nil {
			return fmt.Errorf("could not insert new post like: %w", err)
		}

		// Notify post author if it's a like
		if likeType == 1 {
			var postAuthorID int
			errQuery := tx.QueryRow("SELECT author FROM post WHERE id = ?", postID).Scan(&postAuthorID)
			if errQuery == nil && postAuthorID != userID {
				// Create a notification immediately using the transaction
				_, _ = tx.Exec(`INSERT INTO notifications (user_id, actor_id, post_id, message) VALUES (?, ?, ?, ?)`, postAuthorID, userID, postID, "a aimé votre post")
			}
		}
	}

	return tx.Commit()
}

// LikeComment applies a like or dislike to a comment from a user.
func LikeComment(userID, commentID, likeType int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM comment_likes WHERE user_id = ? AND comment_id = ?`, userID, commentID)
	if err != nil {
		return fmt.Errorf("could not remove existing comment like: %w", err)
	}

	if likeType == 1 || likeType == -1 {
		_, err = tx.Exec(`INSERT INTO comment_likes (user_id, comment_id, type) VALUES (?, ?, ?)`, userID, commentID, likeType)
		if err != nil {
			return fmt.Errorf("could not insert new comment like: %w", err)
		}

		// Notify comment author if it's a like
		if likeType == 1 {
			var commentAuthorID int
			var postID int
			errQuery := tx.QueryRow("SELECT user_id, post_id FROM post_message WHERE id = ?", commentID).Scan(&commentAuthorID, &postID)
			if errQuery == nil && commentAuthorID != userID {
				_, _ = tx.Exec(`INSERT INTO notifications (user_id, actor_id, post_id, message) VALUES (?, ?, ?, ?)`, commentAuthorID, userID, postID, "a aimé votre commentaire")
			}
		}
	}

	return tx.Commit()
}

// GetCommentByID retrieves a single comment by its ID, including like/dislike counts.
func GetCommentByID(id int, currentUserID int) (*Comment, error) {
	comment := &Comment{}
	query := `
		SELECT
			c.id, c.user_id, c.post_id, c.rep_id, c.text, c.timestamp, u.username,
			COALESCE(SUM(CASE WHEN cl.type = 1 THEN 1 ELSE 0 END), 0) AS likes,
			COALESCE(SUM(CASE WHEN cl.type = -1 THEN 1 ELSE 0 END), 0) AS dislikes,
			COALESCE((SELECT type FROM comment_likes WHERE user_id = ? AND comment_id = c.id), 0) AS user_choice
		FROM post_message c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN comment_likes cl ON c.id = cl.comment_id
		WHERE c.id = ?
		GROUP BY c.id`

	err := db.QueryRow(query, currentUserID, id).Scan(
		&comment.ID, &comment.AuthorID, &comment.PostID, &comment.ParentID,
		&comment.Text, &comment.Timestamp, &comment.AuthorUsername,
		&comment.Likes, &comment.Dislikes, &comment.UserChoice,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("could not get comment by id: %w", err)
	}
	return comment, nil
}

// -- Moderation Functions --

// CreateReport adds a new report to the database.
func CreateReport(report *Report) error {
	query := `INSERT INTO reports (reporter_id, content_id, content_type, reason) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, report.ReporterID, report.ContentID, report.ContentType, report.Reason)
	if err != nil {
		return fmt.Errorf("could not create report: %w", err)
	}
	return nil
}

// GetAllReports retrieves all reports from the database.
func GetAllReports() ([]*Report, error) {
	query := `
		SELECT r.id, r.reporter_id, r.content_id, r.content_type, r.reason, r.created_at, u.username, r.status,
		COALESCE(p.author, pm.user_id) as content_author_id,
		pm.post_id as parent_post_id
		FROM reports r
		JOIN users u ON r.reporter_id = u.id
		LEFT JOIN post p ON r.content_type = 'post' AND r.content_id = p.id
		LEFT JOIN post_message pm ON r.content_type = 'comment' AND r.content_id = pm.id
		ORDER BY r.created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not query reports: %w", err)
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		var report Report
		var authorID sql.NullInt64     // Use NullInt64 to handle potential NULLs from LEFT JOIN
		var parentPostID sql.NullInt64 // Handle NULL for posts
		if err := rows.Scan(&report.ID, &report.ReporterID, &report.ContentID, &report.ContentType, &report.Reason, &report.CreatedAt, &report.ReporterName, &report.Status, &authorID, &parentPostID); err != nil {
			return nil, fmt.Errorf("could not scan report: %w", err)
		}
		if authorID.Valid {
			report.ContentAuthorID = int(authorID.Int64)
		}
		if parentPostID.Valid {
			report.ParentPostID = int(parentPostID.Int64)
		} else {
			report.ParentPostID = report.ContentID // Si c'est un post, le post ID est le contentID lui-même
		}
		reports = append(reports, &report)
	}
	return reports, rows.Err()
}

// BanUser marks a user as banned.
func BanUser(userID int) error {
	query := `UPDATE users SET is_banned = 1 WHERE id = ?`
	_, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("could not ban user: %w", err)
	}
	return nil
}

// UpdateReportStatus updates the status of a specific report.
func UpdateReportStatus(reportID int, status string) error {
	query := `UPDATE reports SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, reportID)
	if err != nil {
		return fmt.Errorf("could not update report status: %w", err)
	}
	return nil
}

// UnbanUser marks a user as not banned.
func UnbanUser(userID int) error {
	query := `UPDATE users SET is_banned = 0 WHERE id = ?`
	_, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("could not unban user: %w", err)
	}
	return nil
}

// CreateNotification creates a new notification for a user.
func CreateNotification(userID, actorID int, postID sql.NullInt64, message string) error {
	// Don't notify oneself
	if userID == actorID {
		return nil
	}
	query := `INSERT INTO notifications (user_id, actor_id, post_id, message) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, userID, actorID, postID, message)
	return err
}

// GetUnreadNotifications retrieves unread notifications for a specific user.
func GetUnreadNotifications(userID int) ([]Notification, error) {
	query := `
		SELECT n.id, n.user_id, n.actor_id, u.username as actor_name, n.post_id, n.message, n.is_read, n.created_at
		FROM notifications n
		JOIN users u ON n.actor_id = u.id
		WHERE n.user_id = ? AND n.is_read = 0
		ORDER BY n.created_at DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		var isRead int
		if err := rows.Scan(&n.ID, &n.UserID, &n.ActorID, &n.ActorName, &n.PostID, &n.Message, &isRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.IsRead = isRead == 1
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

// MarkNotificationsAsRead marks all notifications as read for a user.
func MarkNotificationsAsRead(userID int) error {
	query := `UPDATE notifications SET is_read = 1 WHERE user_id = ?`
	_, err := db.Exec(query, userID)
	return err
}

// DeletePost deletes a post and its associated likes and comments.
func DeletePost(postID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	// Cascade delete should handle likes and comments, but explicit is safer
	if _, err := tx.Exec(`DELETE FROM post_likes WHERE post_id = ?`, postID); err != nil {
		return fmt.Errorf("could not delete post likes: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM post_message WHERE post_id = ?`, postID); err != nil {
		return fmt.Errorf("could not delete post comments: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM post WHERE id = ?`, postID); err != nil {
		return fmt.Errorf("could not delete post: %w", err)
	}

	return tx.Commit()
}

// DeleteComment deletes a single comment and its associated likes.
func DeleteComment(commentID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	// Cascade delete should handle likes, but explicit is safer
	if _, err := tx.Exec(`DELETE FROM comment_likes WHERE comment_id = ?`, commentID); err != nil {
		return fmt.Errorf("could not delete comment likes: %w", err)
	}
	// Also delete replies to this comment
	if _, err := tx.Exec(`DELETE FROM post_message WHERE rep_id = ?`, commentID); err != nil {
		return fmt.Errorf("could not delete comment replies: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM post_message WHERE id = ?`, commentID); err != nil {
		return fmt.Errorf("could not delete comment: %w", err)
	}

	return tx.Commit()
}
