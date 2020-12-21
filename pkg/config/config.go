package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

const (
	BucketNameRoot = "sufr"
	BucketNameURL  = "url"
	BucketNameUser = "user"
	BucketNameTag  = "tag"
	DBFileMode     = 0755
	DefaultVersion = "dev"
	// Version        = "1.0.0"
)

var (
	SUFRUserAgent = fmt.Sprintf("Linux:SUFR:v%s", Version)

	BuildTime    string
	BuildGitHash string
	DataDir      string
	DatabaseFile string
	Version      string
)

var (
	DefaultDataDir   = filepath.Join(os.Getenv("HOME"), ".config", "sufr", "data")
	DefaultUserAgent = fmt.Sprintf("Linux:SUFR:v%s", Version)
)

const (
	DefaultBindAddr        = "localhost:8090"
	DefaultDatabaseName    = "sufr.db"
	DefaultSQLDatabaseName = "sufr-sql.db"
	DefaultResultsPerPage  = 40
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
	BindAddr         string   `env:"SUFR_BIND_ADDR"`
	DataDir          string   `env:"SUFR_DATA_DIR"`
	ResultsPerPage   int      `env:"SUFR_RESULTS_PER_PAGE"`
	UserAgent        string   `env:"SUFR_USER_AGENT"`
	DatabaseFilename string   `env:"SUFR_DATABASE_FILENAME"`
	Debug            bool     `env:"SUFR_DEBUG"`
	DatabaseURL      *url.URL `env:"SUFR_DATABASE_URL"`

	// build time information
	Build BuildInfo
}

func (c Config) DatabaseFile() string {
	return filepath.Join(c.DataDir, c.DatabaseFilename)
}

func (c Config) SQLDatabaseFile() string {
	return filepath.Join(c.DataDir, DefaultSQLDatabaseName)
}
