package static

import (
	"embed"
	"net/http"
)

//go:embed public/*
var public embed.FS

func Mount() {
	http.Handle("GET /public/", http.FileServerFS(public))
}

func PublicAsString(name string) string {
	res, err := public.ReadFile("public/" + name)
	if err != nil {
		panic(err)
	}
	return string(res)
}
