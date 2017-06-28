package app

import (
	"html/template"
	"log"
	"net/http"

	"github.com/kyleterry/sufr/data"
	"github.com/oxtoacart/bpool"
)

var bufpool = bpool.NewBufferPool(64)

var templateMap = map[string]*template.Template{
	"url-index":    createTemplate("templates/base.html", "templates/url-index.html"),
	"url-new":      createTemplate("templates/base.html", "templates/url-new.html"),
	"url-view":     createTemplate("templates/base.html", "templates/url-view.html"),
	"url-edit":     createTemplate("templates/base.html", "templates/url-edit.html"),
	"settings":     createTemplate("templates/base.html", "templates/settings.html"),
	"registration": createTemplate("templates/config-base.html", "templates/register.html"),
	"login":        createTemplate("templates/config-base.html", "templates/login.html"),
	"404":          createTemplate("templates/config-base.html", "templates/404.html"),
}

var templateFuncs = template.FuncMap{
	"reverse":    reverse,
	"isyoutube":  isYoutube,
	"youtubevid": youtubevid,
	"isimgur":    isImgur,
	"imgurgal":   imgurgal,
	"newcontext": newcontext,
	"updatePage": updatePage,
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string) error {
	ctx := r.Context()

	t, ok := templateMap[name]
	if !ok {
		log.Fatalf("missing template %s", name)
	}
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	// Avoid partially written responses by writing to a buffer
	templateData := ctx.Value(templateDataKey).(map[string]interface{})

	settings, ok := ctx.Value(settingsKey).(map[string]interface{})
	if ok {
		templateData["Settings"] = settings
	}

	flashes, ok := ctx.Value(flashesKey).(map[string][]interface{})
	if ok {
		templateData["Flashes"] = flashes
	}

	user, ok := ctx.Value(userKey).(*data.User)
	if ok {
		templateData["LoggedIn"] = true
		templateData["User"] = user
	}

	pinnedTags, ok := ctx.Value(pinnedTagsKey).(*data.PinnedTags)
	if ok {
		templateData["PinnedTags"] = pinnedTags
	}

	err := t.ExecuteTemplate(buf, "base", templateData)
	if err != nil {
		return err
	}

	// TODO(kt): Make this changeable when making the API
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
	return nil
}

func createTemplate(files ...string) *template.Template {
	var filebytes = []byte{}
	for _, f := range files {
		filebytes = append(filebytes, MustAsset(f)...)
	}
	tmpl := template.New("*").Funcs(templateFuncs)
	return template.Must(tmpl.Parse(string(filebytes)))
}
