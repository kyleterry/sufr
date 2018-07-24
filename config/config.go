package config

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

const Version = "1.1.1-dev"

const (
	DatabaseName   = "sufr.db"
	BucketNameRoot = "sufr"
	BucketNameURL  = "url"
	BucketNameUser = "user"
	BucketNameTag  = "tag"
	DBFileMode     = 0755

	DefaultPerPage = 40
)

var (
	SUFRUserAgent = fmt.Sprintf("Linux:SUFR:v%s", Version)

	ApplicationBind string
	BuildTime       string
	BuildGitHash    string
	DataDir         string
	DatabaseFile    string
	Debug           bool
)

func New() {
	defaultDataDir := fmt.Sprintf(filepath.Join(os.Getenv("HOME"), ".config", "sufr", "data"))

	flag.StringVar(&ApplicationBind, "bind", "localhost:8090", "Host and port to bind to")
	flag.StringVar(&DataDir, "data-dir", defaultDataDir, "Location to store data in")
	flag.BoolVar(&Debug, "debug", false, "Turn debugging on")

	flag.Parse()

	if _, err := os.Stat(DataDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(DataDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	DatabaseFile = path.Join(DataDir, DatabaseName)
}
