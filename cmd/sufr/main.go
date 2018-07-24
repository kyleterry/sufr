package main

import (
	"log"
	"net/http"

	"github.com/kyleterry/sufr/app"
	"github.com/kyleterry/sufr/config"
	"github.com/kyleterry/sufr/data"
	"github.com/kyleterry/sufr/data/migrations"
)

func main() {
	config.New()
	log.Println("Welcome to SUFR")

	data.MustInit()
	migrations.MustMigrate()

	sufrApp := app.New()

	log.Printf("listening on http://%s", config.ApplicationBind)
	if err := http.ListenAndServe(config.ApplicationBind, sufrApp); err != nil {
		panic(err)
	}
}
