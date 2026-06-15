package utils

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

// templates is a map that will cache all parsed templates.
var templates = make(map[string]*template.Template)

var funcMap = template.FuncMap{
	"initial": func(s string) string {
		if len(s) > 0 {
			return strings.ToUpper(string([]rune(s)[0]))
		}
		return ""
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
}

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
		ts, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			LogFatal(fmt.Sprintf("Failed to parse template %s", name), err)
		}

		// Add the parsed template set to our cache.
		templates[name] = ts
	}
	Log("All templates loaded and cached successfully.")
}

// executeTemplateToBuffer renders a template into memory and returns the result.
// This lets callers set the HTTP status code only after the template has rendered successfully.
func executeTemplateToBuffer(name string, data any) (*bytes.Buffer, error) {
	ts, ok := templates[name]
	if !ok {
		err_msg := fmt.Sprintf("The template %s does not exist.", name)
		LogError(err_msg, nil)
		return nil, errors.New(err_msg)
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "layout.html", data)
	if err != nil {
		LogError(fmt.Sprintf("Error executing template %s", name), err)
		return nil, err
	}

	return buf, nil
}

// RenderTemplate executes a pre-compiled template from the cache, ensuring a fast and
// consistent response.
func RenderTemplate(w http.ResponseWriter, name string, data any) {
	buf, err := executeTemplateToBuffer(name, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		LogError(fmt.Sprintf("Failed to write template %s to response", name), err)
	}
}

// RenderError generates a standardized error page using the error.html template.
func RenderError(statusCode int, title, heading, message string, isAuthenticated bool, w http.ResponseWriter) {
	data := ErrorPageInfo{
		Title:           title,
		Heading:         heading,
		Message:         message,
		Code:            statusCode,
		IsAuthenticated: isAuthenticated,
	}

	buf, err := executeTemplateToBuffer("error.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	_, err = buf.WriteTo(w)
	if err != nil {
		LogError("Failed to write error page to response", err)
	}
}

// PageData is a generic struct for page data. It is currently not used
// by the new template system but is kept for potential future use or compatibility.
type PageData struct {
	Title string
	Body  template.HTML
}

// ErrorPageInfo contains the specific data needed to render the error page.
type ErrorPageInfo struct {
	Code            int
	Heading         string
	Message         string
	Title           string
	IsAuthenticated bool
}
