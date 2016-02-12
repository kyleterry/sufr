//go:generate esc -o ./app/filesystem.go -pkg app static
//go:generate go-bindata -o ./app/template-data.go -pkg app templates
package main

import (
	"log"
	"net/http"

	"github.com/kyleterry/sufr/app"
	"github.com/kyleterry/sufr/config"
)

func main() {
	config.New()
	log.Println("Welcome to SUFR")

	sufrApp := app.New()

	log.Printf("Listening on http://%s", config.ApplicationBind)
	http.ListenAndServe(config.ApplicationBind, sufrApp)
}
