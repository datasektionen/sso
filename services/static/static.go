package static

import (
	"embed"
	"net/http"

	"github.com/datasektionen/logout/pkg/config"
)

//go:embed public/*
var public embed.FS

func Mount() {
	http.Handle("GET /public/", http.FileServerFS(public))
	http.Handle("GET /dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir(config.Config.DistDir))))
}
