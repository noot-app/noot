package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed frontend/*
var embeddedFS embed.FS

// StaticHandler serves the embedded frontend.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(embeddedFS, "frontend")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}
