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

	log.Printf("Listening on %s", config.ApplicationBind)
	http.ListenAndServe(config.ApplicationBind, sufrApp)
}
