package router

import (
	"Forum/internal/db"
	"Forum/pkg/utils"
	"net/http"
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
		Title:      "Home Page",
		Categories: categories,
	}

	utils.RenderTemplate(w, "index.html", data)
}

func demo(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, "demo.html", PageData{Title: "Demo Page"})
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	utils.RenderError(http.StatusForbidden, "403 - Accès refusé", "Erreur 403", "Vous n’avez pas les droits nécessaires pour accéder à cette page.", w)
}
func register(w http.ResponseWriter, r *http.Request){
	data := PageData{Title: "Forum Register"}
	utils.RenderTemplate(w, "register.html", data)
}

func serverError(w http.ResponseWriter, r *http.Request) {
	utils.RenderError(http.StatusInternalServerError, "500 - Erreur serveur", "Erreur 500", "Une erreur interne est survenue. Merci de réessayer plus tard.", w)
}