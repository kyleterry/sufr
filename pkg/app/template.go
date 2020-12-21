package app

import (
	"html/template"
	"log"
	"net/http"

	"github.com/kyleterry/sufr/pkg/data"
	"github.com/kyleterry/sufr/pkg/ui"
	"github.com/oxtoacart/bpool"
	"github.com/shurcooL/httpfs/vfsutil"
)

var bufpool = bpool.NewBufferPool(64)

var templateMap = map[string]*template.Template{
	"url-index":    mustCreateTemplate("templates/base.html", "templates/url-index.html"),
	"url-new":      mustCreateTemplate("templates/base.html", "templates/url-new.html"),
	"url-view":     mustCreateTemplate("templates/base.html", "templates/url-view.html"),
	"url-edit":     mustCreateTemplate("templates/base.html", "templates/url-edit.html"),
	"settings":     mustCreateTemplate("templates/base.html", "templates/settings.html"),
	"registration": mustCreateTemplate("templates/base.html", "templates/register.html"),
	"login":        mustCreateTemplate("templates/base.html", "templates/login.html"),
	"404":          mustCreateTemplate("templates/base.html", "templates/404.html"),
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

func mustCreateTemplate(files ...string) *template.Template {
	var filebytes = []byte{}
	for _, f := range files {
		b, err := vfsutil.ReadFile(ui.NewFileSystem(), f)
		if err != nil {
			log.Fatal(err)
		}

		filebytes = append(filebytes, b...)
	}
	tmpl := template.New("*").Funcs(templateFuncs)
	return template.Must(tmpl.Parse(string(filebytes)))
}
