package app

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/context"
)

func shitbucketImportHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		ctx := context.Get(r, TemplateContext).(map[string]interface{})
		ctx["Title"] = "Import"
		return renderTemplate(w, "shitbucket-import", ctx)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	url := r.PostForm["url"][0]

	session, err := store.Get(r, "flashes")
	if !govalidator.IsURL(url) {
		if err != nil {
			return err
		}
		if url == "" {
			session.AddFlash("URL cannot be blank", "danger")
		} else {
			session.AddFlash(fmt.Sprintf("%s is not a valid URL", url), "danger")
		}
		session.Save(r, w)
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		session.AddFlash(fmt.Sprintf("%s did not return a 200 status code", url), "danger")
		session.Save(r, w)
		http.Redirect(w, r, reverse("shitbucket-import"), http.StatusSeeOther)
		return nil
	}

	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	count, err := shitbucketImporter(content)
	if err != nil {
		session.AddFlash(fmt.Sprintf("There was an error importing: %s", err), "danger")
		session.Save(r, w)
		http.Redirect(w, r, reverse("shitbucket-import"), http.StatusSeeOther)
		return nil
	}

	session.AddFlash(fmt.Sprintf("Successfully added %d URLs from %s", count, url), "success")
	session.Save(r, w)

	http.Redirect(w, r, reverse("shitbucket-import"), http.StatusSeeOther)
	return nil
}
