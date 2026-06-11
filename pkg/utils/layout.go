package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// render parse un fichier template et retourne le résultat sous forme de string.
func Render(fileName string, data any) string {
	pathString := fmt.Sprintf("template/%v.html", fileName)

	t, err := template.ParseFiles(pathString)
	if err != nil {
		LogError(fmt.Sprintf("Error parsing template %s", fileName), err)
		return "Error rendering template"
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		LogError(fmt.Sprintf("Error executing template %s", fileName), err)
		return "Error executing template"
	}
	return buf.String()
}

// executeHTML parse et exécute un fichier template spécifique directement dans le ResponseWriter.
func executeHTML(fileName string, data any, w http.ResponseWriter) {
	pathString := fmt.Sprintf("template/%v.html", fileName)

	t, err := template.ParseFiles(pathString)
	if err != nil {
		LogError(fmt.Sprintf("Error parsing template %s", fileName), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		LogError(fmt.Sprintf("Error executing template %s", fileName), err)
	}
}

// renderFile génère la page complète (avec layout Header/Footer) et l'envoie au client.
func RenderFile(title string, content string, w http.ResponseWriter) {
	data := PageData{
		Title: title,
		Body:  template.HTML(content),
	}

	executeHTML("layout", data, w)
}

// RenderError génère une page d'erreur en utilisant le template error.html.
func RenderError(statusCode int, title, heading, message string, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	content := Render("error", ErrorPageInfo{
		Code:    statusCode,
		Heading: heading,
		Message: message,
	})

	RenderFile(title, content, w)
}

// PageData contient les données de base pour le rendu d'une page HTML (titre, contenu, header, footer).
type PageData struct {
	Title string
	Body  template.HTML
}

// ErrorPageInfo contient les données de base pour le rendu d'une page Error (code d'erreur (404/403/500), le message d'erreur).
type ErrorPageInfo struct {
	Code    int
	Heading string
	Message string
}
