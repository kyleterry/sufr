package app

import "net/http"

func shitbucketImportHandler(w http.ResponseWriter, r *http.Request) error {
	renderTemplate(w, "shitbucket-import", map[string]interface{}{
		"ActiveTab": "imports",
	})
	return nil
}
