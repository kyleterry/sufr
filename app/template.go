package app

import (
	"html/template"
	"log"
	"net/http"
)

var templateMap = map[string]*template.Template{
	"url-index":         createTemplate("templates/base.html", "templates/url-index.html"),
	"url-new":           createTemplate("templates/base.html", "templates/url-new.html"),
	"url-view":          createTemplate("templates/base.html", "templates/url-view.html"),
	"url-edit":          createTemplate("templates/base.html", "templates/url-edit.html"),
	"tag-index":         createTemplate("templates/base.html", "templates/tag-index.html"),
	"tag-view":          createTemplate("templates/base.html", "templates/tag-view.html"),
	"shitbucket-import": createTemplate("templates/base.html", "templates/shitbucket-import.html"),
}

var templateFuncs = template.FuncMap{
	"reverse": reverse,
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	t, ok := templateMap[name]
	if !ok {
		log.Fatalf("missing template %s", name)
	}
	return t.ExecuteTemplate(w, "base", data)
}

func createTemplate(files ...string) *template.Template {
	tmpl := template.New("*").Funcs(templateFuncs)
	return template.Must(tmpl.ParseFiles(files...))
}