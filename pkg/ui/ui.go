package ui

import "net/http"

func Open(path string) (http.File, error) {
	return assets.Open(path)
}

func NewFileSystem() http.FileSystem {
	return assets
}
