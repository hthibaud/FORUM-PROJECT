package utils

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

// templates is a map that will cache all parsed templates.
var templates = make(map[string]*template.Template)

// LoadTemplates finds all page templates, component templates, and the main layout,
// parses them together, and caches them for reuse. This function should be called
// once at application startup.
func LoadTemplates() {
	// Find all our "page" templates (e.g., index.html, about.html)
	pages, err := filepath.Glob("template/*.html")
	if err != nil {
		LogFatal("Failed to find page templates", err)
	}

	// Find all our reusable "component" templates
	components, err := filepath.Glob("template/components/*.html")
	if err != nil {
		LogFatal("Failed to find component templates", err)
	}

	// For each page, we create a template set that includes the page itself,
	// the main layout, and all available components.
	for _, page := range pages {
		name := filepath.Base(page) // e.g., "index.html"

		// Combine the layout, the current page, and all components into one list of files to parse.
		files := append([]string{"template/layout.html", page}, components...)

		// Create and parse the template set.
		// We use New(name) to be able to refer to this specific template set later.
		ts, err := template.New(name).ParseFiles(files...)
		if err != nil {
			LogFatal(fmt.Sprintf("Failed to parse template %s", name), err)
		}

		// Add the parsed template set to our cache.
		templates[name] = ts
	}
	Log("All templates loaded and cached successfully.")
}

// RenderTemplate executes a pre-compiled template from the cache, ensuring a fast and
// consistent response.
func RenderTemplate(w http.ResponseWriter, name string, data any) {
	// Retrieve the requested template from the cache.
	ts, ok := templates[name]
	if !ok {
		// This indicates a developer error (e.g., a typo in the template name).
		err_msg := fmt.Sprintf("The template %s does not exist.", name)
		LogError(err_msg, nil)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the "layout" template. The content of the specific page
	// will be rendered inside it thanks to the `{{template "page" .}}` block.
	err := ts.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		LogError(fmt.Sprintf("Error executing template %s", name), err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RenderError generates a standardized error page using the error.html template.
func RenderError(statusCode int, title, heading, message string, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	data := ErrorPageInfo{
		Title:   title,
		Heading: heading,
		Message: message,
		Code:    statusCode,
	}
	// The data is passed to error.html which is then rendered inside layout.html.
	RenderTemplate(w, "error.html", data)
}

// PageData is a generic struct for page data. It is currently not used
// by the new template system but is kept for potential future use or compatibility.
type PageData struct {
	Title string
	Body  template.HTML
}

// ErrorPageInfo contains the specific data needed to render the error page.
type ErrorPageInfo struct {
	Code    int
	Heading string
	Message string
	Title   string
}
