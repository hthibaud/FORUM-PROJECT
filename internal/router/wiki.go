package router

import (
	"Forum/pkg/utils"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	utils.RenderFile("Forum Accueil", utils.Render("index", nil), w)
}
