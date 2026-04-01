package handler

import (
	"context"
	"merceria/internal/util"
	"net/http"
	"os"
	"text/template"
)

func LogoutHandler(ctx context.Context, fs *os.Root) http.HandlerFunc {
	const name = "logout.html"
	data := ReloadingFile(ctx, fs, name)
	return func(w http.ResponseWriter, r *http.Request) {
		t := util.Must(template.New("").Parse(string(util.Must(data()))))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t.Execute(w, nil)
	}
}
