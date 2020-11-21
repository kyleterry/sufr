package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/joeshaw/envdecode"
	"github.com/kyleterry/sufr/pkg/app"
	"github.com/kyleterry/sufr/pkg/config"
	"github.com/kyleterry/sufr/pkg/data"
	"github.com/kyleterry/sufr/pkg/data/migrations"
)

func main() {
	cfg := &config.Config{}

	config.SetBuildInfo(cfg)

	if err := envdecode.Decode(cfg); err != nil {
		if err != envdecode.ErrNoTargetFieldsAreSet {
			log.Fatal(err)
		}
	}

	config.SetDefaults(cfg)

	flag.StringVar(&cfg.BindAddr, "bind", cfg.BindAddr, "Host and port to bind to")
	flag.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Location to store data in")
	flag.IntVar(&cfg.ResultsPerPage, "results-per-page", cfg.ResultsPerPage, "Results to display per page")
	flag.BoolVar(&cfg.Debug, "debug", cfg.Debug, "Turn debugging on")

	flag.Parse()

	data.MustInit(cfg)
	migrations.MustMigrate(cfg)

	sufrApp := app.New(cfg)

	log.Printf("listening on http://%s", cfg.BindAddr)
	if err := http.ListenAndServe(cfg.BindAddr, sufrApp); err != nil {
		panic(err)
	}
}
