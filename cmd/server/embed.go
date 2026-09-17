package main

import (
	"io/fs"
	"net/http"
	"strings"
)

func embeddedFileHandler(staticFS fs.FS) http.Handler {
	distFS, err := fs.Sub(staticFS, "dist")
	if err != nil {
		distFS = staticFS
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/" {
			path = "/index.html"
		}

		if _, err := distFS.(fs.ReadFileFS).ReadFile(strings.TrimPrefix(path, "/")); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
