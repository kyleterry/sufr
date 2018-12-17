package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const Version = "1.1.1-dev"

const (
	BucketNameRoot = "sufr"
	BucketNameURL  = "url"
	BucketNameUser = "user"
	BucketNameTag  = "tag"
	DBFileMode     = 0755
)

var (
	SUFRUserAgent = fmt.Sprintf("Linux:SUFR:v%s", Version)

	BuildTime    string
	BuildGitHash string
	DataDir      string
	DatabaseFile string
)

var (
	DefaultDataDir   = fmt.Sprintf(filepath.Join(os.Getenv("HOME"), ".config", "sufr", "data"))
	DefaultUserAgent = fmt.Sprintf("Linux:SUFR:v%s", Version)
)

const (
	DefaultBindAddr       = "localhost:8090"
	DefaultDatabaseName   = "sufr.db"
	DefaultResultsPerPage = 40
)

type BuildInfo struct {
	Time    string
	GitHash string
}

func SetBuildInfo(cfg *Config) {
	cfg.Build = BuildInfo{
		Time:    BuildTime,
		GitHash: BuildGitHash,
	}
}

func SetDefaults(cfg *Config) {
	if cfg.BindAddr == "" {
		cfg.BindAddr = DefaultBindAddr
	}

	if cfg.DataDir == "" {
		cfg.DataDir = DefaultDataDir
	}

	if cfg.ResultsPerPage == 0 {
		cfg.ResultsPerPage = DefaultResultsPerPage
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}

	if cfg.DatabaseFilename == "" {
		cfg.DatabaseFilename = DefaultDatabaseName
	}
}

type Config struct {
	BindAddr         string `env:"SUFR_BIND_ADDR"`
	DataDir          string `env:"SUFR_DATA_DIR"`
	ResultsPerPage   int    `env:"SUFR_RESULTS_PER_PAGE"`
	UserAgent        string `env:"SUFR_USER_AGENT"`
	DatabaseFilename string `env:"SUFR_DATABASE_FILENAME"`
	Debug            bool   `env:"SUFR_DEBUG"`

	// build time information
	Build BuildInfo
}

func (c Config) DatabaseFile() string {
	return filepath.Join(c.DataDir, c.DatabaseFilename)
}
