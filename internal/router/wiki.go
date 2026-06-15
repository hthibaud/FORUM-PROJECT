package router

import (
	"Forum/internal/db"
	"Forum/internal/session"
	"Forum/pkg/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

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

	data := PageData{
		Title:           "Home Page",
		Categories:      categories,
		IsAuthenticated: session.IsAuthenticated(r),
	}

	utils.RenderTemplate(w, "index.html", data)
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	utils.RenderError(http.StatusForbidden, "403 - Accès refusé", "Erreur 403", "Vous n’avez pas les droits nécessaires pour accéder à cette page.", session.IsAuthenticated(r), w)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	utils.RenderError(http.StatusNotFound, "404 - Page non trouvée", "Erreur 404", "La page que vous recherchez n'existe pas.", session.IsAuthenticated(r), w)
}

func register(w http.ResponseWriter, r *http.Request) {
	utils.Debug("Register page accessed")
	data := PageData{Title: "Forum Register"}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		utils.Debug("Registering user: " + username)

		// Validation simple
		if username == "" || email == "" || password == "" {
			utils.Debug("Validation failed: fields missing")
			data.Message = "Tous les champs sont requis"
			utils.RenderTemplate(w, "register.html", data)
			return
		}

		// Vérifier si l'utilisateur existe déjà
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

		// Créer l'utilisateur
		err = db.CreateUser(username, password, email)
		if err != nil {
			utils.LogError("Erreur lors de la création de l'utilisateur", err)
			serverError(w, r)
			return
		}
		utils.Debug("User created successfully: " + username)

		// Connecter l'utilisateur
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
	utils.RenderError(http.StatusInternalServerError, "500 - Erreur serveur", "Erreur 500", "Une erreur interne est survenue. Merci de réessayer plus tard.", session.IsAuthenticated(r), w)
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

	data := PageData{
		Title:            selectedCategory.Name,
		Categories:       categories,
		SelectedCategory: selectedCategory,
		Posts:            posts,
		Notifications:    []db.Notification{},
		IsAuthenticated:  session.IsAuthenticated(r),
	}
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

	// Convert idStr to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.LogError("Invalid post id", err)
		http.NotFound(w, r)
		return
	}

	post, err := db.GetPostByID(id)
	if err != nil {
		utils.LogError("Error retrieving post", err)
		serverError(w, r)
		return
	}

	if post == nil {
		http.NotFound(w, r)
		return
	}

	comments, err := db.GetCommentsByPostID(id)
	if err != nil {
		utils.LogError("Error retrieving comments", err)
		serverError(w, r)
		return
	}

	data := PageData{
		Title:           post.Title,
		IsAuthenticated: session.IsAuthenticated(r),
		Post:            post,
		Comments:        comments,
	}

	utils.RenderTemplate(w, "post.html", data)
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

		data := PageData{
			Title:           "Create a new post",
			IsAuthenticated: true,
			Categories:      categories,
		}
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

	// Create comment in database
	err = db.CreateComment(user.ID, postID, sql.NullInt64{}, commentText)
	if err != nil {
		utils.LogError("Failed to create comment in database", err)
		serverError(w, r)
		return
	}
	utils.Debug(fmt.Sprintf("Successfully created comment for post %d by user %d", postID, user.ID))

	http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
}
