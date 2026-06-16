package router

import (
	"Forum/internal/db"
	"Forum/internal/session"
	"Forum/pkg/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// newAuthenticatedPageData creates a PageData struct and populates it with user
// information if the user is authenticated.
func newAuthenticatedPageData(r *http.Request) PageData {
	data := PageData{
		IsAuthenticated: session.IsAuthenticated(r),
	}
	if data.IsAuthenticated {
		user, err := session.GetUserFromSession(r)
		if err == nil && user != nil {
			data.User = *user

			// Load notifications for the authenticated user
			notifications, err := db.GetUnreadNotifications(user.ID)
			if err == nil {
				data.Notifications = notifications
			}
		}
	}
	return data
}

// checkBannedStatus is a middleware that checks if a user is banned.
// If so, it redirects them to the banned page.
func checkBannedStatus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if session.IsAuthenticated(r) {
			user, err := session.GetUserFromSession(r)
			if err == nil && user != nil && user.IsBanned {
				// Allow banned users to log out
				if r.URL.Path == "/logout" {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, "/banned", http.StatusSeeOther)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func bannedPage(w http.ResponseWriter, r *http.Request) {
	data := newAuthenticatedPageData(r)
	data.Title = "Banni"
	utils.RenderTemplate(w, "banned.html", data)
}

func rulesPage(w http.ResponseWriter, r *http.Request) {
	data := newAuthenticatedPageData(r)
	data.Title = "Règles du Forum"
	utils.RenderTemplate(w, "rules.html", data)
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	data := newAuthenticatedPageData(r)
	data.Title = "Nous Contacter"
	utils.RenderTemplate(w, "contact.html", data)
}

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFound(w, r)
		return
	}

	categories, err := db.GetCategories()
	if err != nil {
		utils.LogError("could not get categories", err)
		serverError(w, r)
		return
	}

	recentPosts, err := db.GetRecentPosts(5) // Retrieve the 5 most recent
	if err != nil {
		utils.LogError("could not get recent posts", err)
		serverError(w, r)
		return
	}

	data := newAuthenticatedPageData(r)
	data.Title = "Accueil"
	data.Categories = categories
	data.Posts = recentPosts

	utils.RenderTemplate(w, "index.html", data)
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	categories, err := db.GetCategories()
	if err != nil {
		utils.LogError("could not get categories for forbidden page", err)
	}
	user, _ := session.GetUserFromSession(r)
	utils.RenderError(w, http.StatusForbidden, "403 - Accès refusé", "Erreur 403", "Vous n’avez pas les droits nécessaires pour accéder à cette page.", session.IsAuthenticated(r), categories, user)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	categories, err := db.GetCategories()
	if err != nil {
		utils.LogError("could not get categories for notFound page", err)
	}
	user, _ := session.GetUserFromSession(r)
	utils.RenderError(w, http.StatusNotFound, "404 - Page non trouvée", "Erreur 404", "La page que vous recherchez n'existe pas.", session.IsAuthenticated(r), categories, user)
}

func register(w http.ResponseWriter, r *http.Request) {
	utils.Debug("Register page accessed")
	data := PageData{Title: "Forum Register"}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		utils.Debug("Registering user: " + username)

		// Simple validation
		if username == "" || email == "" || password == "" {
			utils.Debug("Validation failed: fields missing")
			data.Message = "Tous les champs sont requis"
			utils.RenderTemplate(w, "register.html", data)
			return
		}

		// Check if the user already exists
		existingUser, err := db.GetUserByUsername(username)
		if err != nil {
			utils.LogError("Erreur lors de la vérification de l'utilisateur", err)
			serverError(w, r)
			return
		}
		if existingUser != nil {
			utils.Debug("User already exists: " + username)
			data.Message = "Ce nom d'utilisateur est déjà pris"
			utils.RenderTemplate(w, "register.html", data)
			return
		}

		// Check if the email is already used
		existingEmailUser, err := db.GetUserByEmail(email)
		if err != nil {
			utils.LogError("Erreur lors de la vérification de l'email", err)
			serverError(w, r)
			return
		}
		if existingEmailUser != nil {
			utils.Debug("Email already in use: " + email)
			data.Message = "Cette adresse e-mail est déjà utilisée"
			utils.RenderTemplate(w, "register.html", data)
			return
		}

		// Create the user
		err = db.CreateUser(username, password, email)
		if err != nil {
			utils.LogError("Erreur lors de la création de l'utilisateur", err)
			serverError(w, r)
			return
		}
		utils.Debug("User created successfully: " + username)

		// Sign the user in
		user, err := db.GetUserByUsername(username)
		if err != nil || user == nil {
			utils.LogError("Erreur pour retrouver l'utilisateur après création", err)
			serverError(w, r)
			return
		}

		err = session.CreateSession(w, user.ID, r)
		if err != nil {
			utils.LogError("Erreur lors de la création de la session", err)
			serverError(w, r)
			return
		}
		utils.Debug("Session created for user: " + username)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	utils.RenderTemplate(w, "register.html", data)
}

func serverError(w http.ResponseWriter, r *http.Request) {
	categories, err := db.GetCategories()
	if err != nil {
		utils.LogError("could not get categories for serverError page, sending plain error", err)
	}
	user, _ := session.GetUserFromSession(r)
	utils.RenderError(w, http.StatusInternalServerError, "500 - Erreur serveur", "Erreur 500", "Une erreur interne est survenue. Merci de réessayer plus tard.", session.IsAuthenticated(r), categories, user)
}

func login(w http.ResponseWriter, r *http.Request) {
	utils.Debug("Login page accessed")
	data := PageData{Title: "Forum Login"}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")
		utils.Debug("Attempting to login user: " + username)

		if username == "" || password == "" {
			utils.Debug("Validation failed: fields missing")
			data.Message = "Nom d'utilisateur et mot de passe requis"
			utils.RenderTemplate(w, "login.html", data)
			return
		}

		user, err := db.GetUserByUsername(username)
		if err != nil {
			utils.LogError("Erreur lors de la récupération de l'utilisateur", err)
			serverError(w, r)
			return
		}

		if user == nil || !utils.CheckPasswordHash(password, user.Password) {
			utils.Debug("Invalid credentials for user: " + username)
			data.Message = "Identifiants invalides"
			utils.RenderTemplate(w, "login.html", data)
			return
		}
		if user.IsBanned {
			utils.Debug("Banned user tried to login: " + username)
			data.Message = "Votre compte a été banni."
			utils.RenderTemplate(w, "login.html", data)
			return
		}
		utils.Debug("User authenticated successfully: " + username)

		err = session.CreateSession(w, user.ID, r)
		if err != nil {
			utils.LogError("Erreur lors de la création de la session", err)
			serverError(w, r)
			return
		}
		utils.Debug("Session created for user: " + username)

		http.Redirect(w, r, "/category/1", http.StatusSeeOther)
		return
	}

	utils.RenderTemplate(w, "login.html", data)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func categoryPage(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/category/") {
		http.NotFound(w, r)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/category/")
	if idStr == "" || strings.Contains(idStr, "/") {
		http.NotFound(w, r)
		return
	}

	categoryID, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	categories, err := db.GetCategories()
	if err != nil {
		utils.LogError("could not get categories", err)
		serverError(w, r)
		return
	}

	selectedCategory, err := db.GetCategoryByID(categoryID)
	if err != nil {
		utils.LogError("could not get selected category", err)
		serverError(w, r)
		return
	}
	if selectedCategory == nil {
		http.NotFound(w, r)
		return
	}

	posts, err := db.GetPostsByCategory(categoryID)
	if err != nil {
		utils.LogError("could not get posts for category", err)
		serverError(w, r)
		return
	}

	data := newAuthenticatedPageData(r)
	data.Title = selectedCategory.Name
	data.Categories = categories
	data.SelectedCategory = selectedCategory
	data.Posts = posts
	data.Notifications = []db.Notification{}

	utils.RenderTemplate(w, "general.html", data)
}

func postView(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/comment") && r.Method == http.MethodPost {
		handleCommentSubmission(w, r)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/post/")
	if idStr == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogError("Invalid post id", err)
		http.NotFound(w, r)
		return
	}

	// Get current user, if any
	user, _ := session.GetUserFromSession(r)
	var currentUserID int
	if user != nil {
		currentUserID = user.ID
	}

	post, err := db.GetPostByID(id, currentUserID)
	if err != nil {
		utils.LogError("Error retrieving post", err)
		serverError(w, r)
		return
	}

	if post == nil {
		http.NotFound(w, r)
		return
	}

	comments, err := db.GetCommentsForPost(id, currentUserID)
	if err != nil {
		utils.LogError("Error retrieving comments", err)
		serverError(w, r)
		return
	}

	// Organize comments into a hierarchical structure
	commentMap := make(map[int]*db.Comment)
	for i := range comments {
		commentMap[comments[i].ID] = &comments[i]
	}

	var rootComments []*db.Comment
	for i := range comments {
		comment := &comments[i]
		if comment.ParentID.Valid {
			if parent, ok := commentMap[int(comment.ParentID.Int64)]; ok {
				parent.Replies = append(parent.Replies, comment)
			}
		} else {
			rootComments = append(rootComments, comment)
		}
	}

	categories, err := db.GetCategories()
	if err != nil {
		utils.LogError("could not get categories", err)
		serverError(w, r)
		return
	}

	data := newAuthenticatedPageData(r)
	data.Title = post.Title
	data.Post = *post
	data.Comments = rootComments
	data.Categories = categories

	utils.RenderTemplate(w, "post.html", data)
}

func profile(w http.ResponseWriter, r *http.Request) {
	utils.Debug("Accessing profile page")
	if !session.IsAuthenticated(r) {
		utils.Debug("User not authenticated, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := session.GetUserFromSession(r)
	if err != nil || user == nil {
		utils.LogError("could not get authenticated user", err)
		serverError(w, r)
		return
	}

	categories, err := db.GetCategories()
	if err != nil {
		utils.LogError("could not get categories", err)
		serverError(w, r)
		return
	}

	createdPostsCount := 0
	likedPostsCount := 0
	postedCommentsCount := 0
	var createdPosts []db.Post
	var likedPosts []db.Post
	var postedComments []db.Comment

	for _, category := range categories {
		posts, err := db.GetPostsByCategory(category.ID)
		if err != nil {
			utils.LogError("could not get posts for category", err)
			serverError(w, r)
			return
		}

		for _, post := range posts {
			if post.AuthorID == user.ID {
				createdPostsCount++
				createdPosts = append(createdPosts, post)
			}

			detailedPost, err := db.GetPostByID(post.ID, user.ID)
			if err != nil {
				utils.LogError("could not get detailed post info", err)
				serverError(w, r)
				return
			}
			if detailedPost != nil && detailedPost.UserChoice == 1 {
				likedPostsCount++
				likedPosts = append(likedPosts, *detailedPost)
			}

			comments, err := db.GetCommentsForPost(post.ID, user.ID)
			if err != nil {
				utils.LogError("could not get comments for post", err)
				serverError(w, r)
				return
			}
			for _, comment := range comments {
				if comment.AuthorID == user.ID {
					postedCommentsCount++
					postedComments = append(postedComments, comment)
				}
			}
		}
	}

	data := map[string]any{
		"Title":               "Profil",
		"User":                *user,
		"Categories":          categories,
		"IsAuthenticated":     true,
		"CreatedPostsCount":   createdPostsCount,
		"LikedPostsCount":     likedPostsCount,
		"PostedCommentsCount": postedCommentsCount,
		"CreatedPosts":        createdPosts,
		"LikedPosts":          likedPosts,
		"PostedComments":      postedComments,
	}

	utils.RenderTemplate(w, "profile.html", data)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	utils.Debug("Accessing create post page")
	if !session.IsAuthenticated(r) {
		utils.Debug("User not authenticated, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		utils.Debug("Displaying create post form")
		categories, err := db.GetCategories()
		if err != nil {
			utils.LogError("could not get categories", err)
			serverError(w, r)
			return
		}

		data := newAuthenticatedPageData(r)
		data.Title = "Create a new post"
		data.Categories = categories

		utils.RenderTemplate(w, "create_post.html", data)
		return
	}

	if r.Method == http.MethodPost {
		utils.Debug("Handling post creation form submission")
		err := r.ParseForm()
		if err != nil {
			serverError(w, r)
			return
		}

		title := r.FormValue("title")
		text := r.FormValue("content")
		categoryIDStr := r.FormValue("category_id")
		utils.Debug(fmt.Sprintf("Form values: title='%s', category_id='%s'", title, categoryIDStr))

		if title == "" || text == "" || categoryIDStr == "" {
			utils.Debug("Missing fields in create post form")
			http.Redirect(w, r, "/post/create?error=missing_fields", http.StatusSeeOther)
			return
		}

		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			utils.Debug("Invalid category ID")
			http.Redirect(w, r, "/post/create?error=invalid_category", http.StatusSeeOther)
			return
		}

		user, err := session.GetUserFromSession(r)
		if err != nil {
			// This case is for database errors etc.
			utils.LogError("an unexpected error occurred getting user from session", err)
			serverError(w, r)
			return
		}
		if user == nil {
			// This case is for no session/invalid session
			utils.Debug("GetUserFromSession returned no user, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		utils.Debug(fmt.Sprintf("Creating post for user: %s (ID: %d)", user.Username, user.ID))

		postID, err := db.CreatePost(user.ID, categoryID, title, text)
		if err != nil {
			utils.LogError("could not create post", err)
			serverError(w, r)
			return
		}
		utils.Debug(fmt.Sprintf("Post created successfully with ID: %d", postID))

		http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
	}
}

func handleCommentSubmission(w http.ResponseWriter, r *http.Request) {
	utils.Debug("Handling comment submission")

	// Extract Post ID from URL
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/post/"), "/comment")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogError("Invalid post ID in comment submission", err)
		http.NotFound(w, r)
		return
	}
	utils.Debug(fmt.Sprintf("Comment submission for post ID: %d", postID))

	// Check if user is authenticated
	user, err := session.GetUserFromSession(r)
	if err != nil {
		utils.LogError("Error getting user from session for comment", err)
		serverError(w, r)
		return
	}
	if user == nil {
		utils.Debug("Unauthenticated user tried to comment, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	utils.Debug(fmt.Sprintf("User '%s' (ID: %d) is submitting a comment", user.Username, user.ID))

	// Parse form and get comment text
	if err := r.ParseForm(); err != nil {
		utils.LogError("Error parsing comment form", err)
		serverError(w, r)
		return
	}
	commentText := r.FormValue("comment")
	if commentText == "" {
		utils.Debug("Comment submission is empty")
		http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
		return
	}

	parentIDStr := r.FormValue("parent_id")
	var parentID sql.NullInt64
	if parentIDStr != "" {
		pID, err := strconv.Atoi(parentIDStr)
		if err == nil {
			parentID = sql.NullInt64{Int64: int64(pID), Valid: true}
		}
	}

	// Create comment in database
	comment := db.Comment{
		AuthorID:  user.ID,
		PostID:    postID,
		ParentID:  parentID,
		Text:      commentText,
		Timestamp: time.Now(),
	}
	err = db.CreateComment(comment)
	if err != nil {
		utils.LogError("Failed to create comment in database", err)
		serverError(w, r)
		return
	}
	utils.Debug(fmt.Sprintf("Successfully created comment for post %d by user %d", postID, user.ID))

	http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
}

func handlePostLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := session.GetUserFromSession(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		PostID int `json:"post_id"`
		Type   int `json:"type"` // 1 for like, -1 for dislike, 0 to remove
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := db.LikePost(user.ID, req.PostID, req.Type); err != nil {
		utils.LogError("could not like post", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch updated counts
	post, err := db.GetPostByID(req.PostID, user.ID)
	if err != nil || post == nil {
		utils.LogError("could not get post by id after like", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"likes":      post.Likes,
		"dislikes":   post.Dislikes,
		"userChoice": post.UserChoice,
	})
}

func handleCommentLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := session.GetUserFromSession(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CommentID int `json:"comment_id"`
		Type      int `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if err := db.LikeComment(user.ID, req.CommentID, req.Type); err != nil {
		utils.LogError("could not like comment", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// To send back the updated counts, we need a way to get a single comment's data.
	// This function needs to be created in db/db.go
	comment, err := db.GetCommentByID(req.CommentID, user.ID)
	if err != nil || comment == nil {
		utils.LogError("could not get comment by id after like", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"likes":      comment.Likes,
		"dislikes":   comment.Dislikes,
		"userChoice": comment.UserChoice,
	})
}

func readNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := session.GetUserFromSession(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := db.MarkNotificationsAsRead(user.ID); err != nil {
		utils.LogError("could not mark notifications as read", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// -- Moderation Handlers --

func isModerator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := session.GetUserFromSession(r)
		if err != nil {
			utils.LogError("Error getting user from session", err)
			serverError(w, r)
			return
		}
		if user == nil || (user.Role != "moderator" && user.Role != "admin") {
			forbidden(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func reportContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := session.GetUserFromSession(r)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	if err != nil {
		serverError(w, r)
		return
	}

	contentType := r.FormValue("content_type")
	contentIDStr := r.FormValue("content_id")
	reason := r.FormValue("reason")

	contentID, err := strconv.Atoi(contentIDStr)
	if err != nil {
		http.Error(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	report := &db.Report{
		ReporterID:  user.ID,
		ContentID:   contentID,
		ContentType: contentType,
		Reason:      reason,
	}

	err = db.CreateReport(report)
	if err != nil {
		utils.LogError("could not create report", err)
		serverError(w, r)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func moderationPage(w http.ResponseWriter, r *http.Request) {
	reports, err := db.GetAllReports()
	if err != nil {
		utils.LogError("could not get reports", err)
		serverError(w, r)
		return
	}

	data := newAuthenticatedPageData(r)
	data.Title = "Moderation"
	data.Reports = reports

	utils.RenderTemplate(w, "moderation.html", data)
}

func banUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/moderation/ban/")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = db.BanUser(userID)
	if err != nil {
		utils.LogError("could not ban user", err)
		serverError(w, r)
		return
	}

	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func handleDeleteReport(w http.ResponseWriter, r *http.Request) {
	reportIDStr := strings.TrimPrefix(r.URL.Path, "/moderation/handle/delete/")
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// This is not efficient, but it's simple. A better way would be to get the report by its ID.
	// I will assume this is acceptable for now.
	reports, err := db.GetAllReports()
	if err != nil {
		serverError(w, r)
		return
	}

	var report *db.Report
	for _, r := range reports {
		if r.ID == reportID {
			report = r
			break
		}
	}

	if report == nil {
		http.NotFound(w, r)
		return
	}

	if report.ContentType == "post" {
		err = db.DeletePost(report.ContentID)
	} else if report.ContentType == "comment" {
		err = db.DeleteComment(report.ContentID)
	}

	if err != nil {
		utils.LogError("could not delete content from report", err)
		serverError(w, r)
		return
	}

	err = db.UpdateReportStatus(reportID, "handled")
	if err != nil {
		utils.LogError("could not update report status", err)
		// The content was deleted, but the status update failed. Log and redirect.
	}

	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func handleDismissReport(w http.ResponseWriter, r *http.Request) {
	reportIDStr := strings.TrimPrefix(r.URL.Path, "/moderation/handle/dismiss/")
	reportID, err := strconv.Atoi(reportIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = db.UpdateReportStatus(reportID, "handled")
	if err != nil {
		utils.LogError("could not dismiss report", err)
		serverError(w, r)
		return
	}

	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func unbanUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/moderation/unban/")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = db.UnbanUser(userID)
	if err != nil {
		utils.LogError("could not unban user", err)
		serverError(w, r)
		return
	}

	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/moderation/delete/post/")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = db.DeletePost(postID)
	if err != nil {
		utils.LogError("could not delete post", err)
		serverError(w, r)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func deleteComment(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/moderation/delete/comment/")
	commentID, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = db.DeleteComment(commentID)
	if err != nil {
		utils.LogError("could not delete comment", err)
		serverError(w, r)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
