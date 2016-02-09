package app

import (
	"net/http"

	"github.com/gorilla/context"
)

func shitbucketImportHandler(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Get(r, TemplateContext).(map[string]interface{})
	ctx["Title"] = "Import"

	renderTemplate(w, "shitbucket-import", ctx)
	return nil
}
