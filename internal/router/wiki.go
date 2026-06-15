package router

import (
	"Forum/internal/db"
	"Forum/internal/session"
	"Forum/pkg/utils"
	"net/http"
	"strconv"
	"strings"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
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
	utils.RenderError(http.StatusForbidden, "403 - Accès refusé", "Erreur 403", "Vous n’avez pas les droits nécessaires pour accéder à cette page.", w)
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
	utils.RenderError(http.StatusInternalServerError, "500 - Erreur serveur", "Erreur 500", "Une erreur interne est survenue. Merci de réessayer plus tard.", w)
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
		IsAuthenticated:  session.IsAuthenticated(r),
	}
	utils.RenderTemplate(w, "general.html", data)
}
