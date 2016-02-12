package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const Version = "1.0.0"

const (
	DatabaseName   = "sufr.db"
	BucketNameRoot = "sufr"
	BucketNameURL  = "url"
	BucketNameUser = "user"
	BucketNameTag  = "tag"
	DBFileMode     = 0755
)

var (
	ApplicationBind string
	BuildTime       string
	BuildGitHash    string
	DataDir         string
	DatabaseFile    string
	TemplateDir     string
	StaticDir       string
	RootDir         string
	Debug           bool

	ErrDatabaseAlreadyOpen = errors.New("Database is already open")
	ErrKeyNotFound         = errors.New("Key doesn't exist in DB")
)

func New() {
	RootDir = findWorkingDir()

	flag.StringVar(&ApplicationBind, "bind", "localhost:8090", "Host and port to bind to")

	//TODO(kt): handle windows configuration dir
	defaultDataDir := fmt.Sprintf(path.Join(os.Getenv("HOME"), ".config", "sufr", "data"))
	flag.StringVar(&DataDir, "data-dir", defaultDataDir, "Location to store data in")

	defaultTemplateDir := path.Join(RootDir, "templates")
	flag.StringVar(&TemplateDir, "template-dir", defaultTemplateDir, "Location where templates are stored")

	defaultStaticDir := path.Join(RootDir, "static")
	flag.StringVar(&StaticDir, "static-dir", defaultStaticDir, "Location where static assets are stored")

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

func execPath() (string, error) {
	f, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Abs(f)
}

// Tries to find the working directory (where templates and static files are).
// If it can't fetch it from the config file, it will just use the execution path.
// Returns a string
func findWorkingDir() string {
	path, err := execPath()
	if err != nil {
		panic("Cannot find exec path")
	}
	i := strings.LastIndex(path, "/")
	return path[:i]
}
