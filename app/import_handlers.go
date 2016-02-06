package app

import "net/http"

func (s *Sufr) shitbucketImportHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "shitbucket-import", map[string]interface{}{
		"ActiveTab": "imports",
	})
}
