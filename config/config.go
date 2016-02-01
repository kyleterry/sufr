package config

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	ApplicationBind string
	DataDir         string
	TemplateDir     string
	StaticDir       string
	RootDir         string
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

	flag.Parse()
}

func ExecPath() (string, error) {
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
	path, err := ExecPath()
	if err != nil {
		panic("Cannot find exec path")
	}
	i := strings.LastIndex(path, "/")
	return path[:i]
}
