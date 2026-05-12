package handler

import (
	"context"
	"fmt"
	"html/template"
	"log"
	apperr "merceria/internal/model/error"
	"merceria/internal/util"
	"net/http"
	"os"
)

type Handler func(r *http.Request) (func(w http.ResponseWriter), error)

func NewAdapter(ctx context.Context, fs *os.Root) (func(h Handler) http.HandlerFunc, error) {
	const name = "templates/error.html"
	Render, err := ErrorRenderer(ctx, fs, name)
	if err != nil {
		return nil, fmt.Errorf("building error renderer: %w", err)
	}

	return func(h Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			Success, err := h(r)
			if err == nil && Success == nil {
				err = fmt.Errorf("path: %s", r.URL.String())
				err = apperr.WithCode(err, "EMTPY_SUCCESS")
				err = apperr.WithMessage(err, "the received handler returned an empty success")
			}
			if err != nil {
				Render(w, err)
				return
			}
			Success(w)
		}
	}, nil
}

func ErrorRenderer(ctx context.Context, fs *os.Root, path string) (func(w http.ResponseWriter, err error), error) {
	Load := func() (*template.Template, error) {
		data, err := fs.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
		tpl, err := template.New(path).Parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("parsing template: %w", err)
		}
		return tpl, nil
	}

	tpl, err := Load()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	go func() {
		for range util.ShittyFsNotify(ctx, fs, path) {
			log.Printf("%s changed: reloading", path)
			t, err := Load()
			if err != nil {
				log.Printf("%s: failed to reload data: %s", path, err)
				continue
			}
			tpl = t
		}
	}()

	return func(w http.ResponseWriter, err error) {
		type DetailedError struct {
			Code    string
			Title   string
			Message string
			Inner   error
		}
		render := &DetailedError{
			Code:    apperr.Code(err),
			Title:   apperr.Title(err),
			Message: apperr.Message(err),
		}
		if err := apperr.Public(err); err != nil {
			render.Inner = err
		}
		if status := apperr.Status(err); status != 0 {
			w.WriteHeader(status)
		}
		tpl.Execute(w, render)
	}, nil
}
