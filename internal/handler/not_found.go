package handler

import (
	"net/http"
	"os"
)

func NotFound(fs *os.Root) http.HandlerFunc {
	const name = "404.html"

	data, err := fs.ReadFile(name)
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write(data)
	}
}
