//go:generate go-bindata -o ./app/bindata.go -pkg app templates/... static/...
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
	if err := http.ListenAndServe(config.ApplicationBind, sufrApp); err != nil {
		panic(err)
	}
}
