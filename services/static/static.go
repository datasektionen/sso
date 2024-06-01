package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed files/*
var files embed.FS

func Mount() {
	fs, err := fs.Sub(files, "files")
	if err != nil {
		panic(err)
	}
	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(fs)))
}
