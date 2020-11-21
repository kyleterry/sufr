// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
)

var assets = union.New(map[string]http.FileSystem{
	"/static":    http.Dir("static"),
	"/templates": http.Dir("templates"),
})

func main() {
	err := vfsgen.Generate(assets, vfsgen.Options{
		Filename:    "build_assets.go",
		PackageName: "ui",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
