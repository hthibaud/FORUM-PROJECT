package router

import (
	"Forum/pkg/utils"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		utils.RenderError(http.StatusNotFound, "404 - Page introuvable", "Erreur 404", "La page demandée n'existe pas.", w)
		return
	}
	utils.RenderFile("Forum Accueil", utils.Render("index", nil), w)
}

func demo(w http.ResponseWriter, r *http.Request) {
	utils.RenderFile("Forum Demo", utils.Render("demo", nil), w)
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	utils.RenderError(http.StatusForbidden, "403 - Accès refusé", "Erreur 403", "Vous n’avez pas les droits nécessaires pour accéder à cette page.", w)
}
func register(w http.ResponseWriter, r *http.Request) {
	utils.RenderFile("Forum Register", utils.Render("register", nil), w)
}

func login(w http.ResponseWriter, r *http.Request) {
	utils.RenderFile("Forum Login", utils.Render("login", nil), w)
}

func serverError(w http.ResponseWriter, r *http.Request) {
	utils.RenderError(http.StatusInternalServerError, "500 - Erreur serveur", "Erreur 500", "Une erreur interne est survenue. Merci de réessayer plus tard.", w)
}
