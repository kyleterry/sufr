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
	"settings":          createTemplate("templates/base.html", "templates/settings.html"),
	"shitbucket-import": createTemplate("templates/base.html", "templates/shitbucket-import.html"),
	"registration":      createTemplate("templates/config-base.html", "templates/register.html"),
	"login":             createTemplate("templates/config-base.html", "templates/login.html"),
}

var templateFuncs = template.FuncMap{
	"reverse":    reverse,
	"isyoutube":  isYoutube,
	"youtubevid": youtubevid,
	"newcontext": newcontext,
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	t, ok := templateMap[name]
	if !ok {
		log.Fatalf("missing template %s", name)
	}
	return t.ExecuteTemplate(w, "base", data)
}

func createTemplate(files ...string) *template.Template {
	var filebytes = []byte{}
	for _, f := range files {
		filebytes = append(filebytes, MustAsset(f)...)
	}
	tmpl := template.New("*").Funcs(templateFuncs)
	return template.Must(tmpl.Parse(string(filebytes)))
}
