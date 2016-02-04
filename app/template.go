package app

import (
	"html/template"
	"net/http"
)

var tempmap = map[string]*template.Template{
	"url-index": createTemplate("templates/base.html", "templates/url-index.html"),
}

func renderTemplate(w http.ResponseWriter, t string, data interface{}) error {
	return tempmap[t].ExecuteTemplate(w, "base", data)
}

func createTemplate(files ...string) *template.Template {
	return template.Must(template.New("*").ParseFiles(files...))
}
