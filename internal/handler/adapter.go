package handler

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"merceria/internal/model/apperr"
	"merceria/internal/util"
	"net/http"
	"os"
)

type HandlerFunc = func(w http.ResponseWriter, r *http.Request) error

func Adapter(Err ErrorRenderer) func(fn HandlerFunc) http.HandlerFunc {
	return func(fn HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			err := fn(w, r)
			if err != nil {
				Err(w, r, err)
				return
			}
		}
	}
}

type ErrorRenderer = func(w http.ResponseWriter, r *http.Request, err error)

func NewErrorRenderer(ctx context.Context, fs *os.Root, path string) (ErrorRenderer, error) {
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

	return func(w http.ResponseWriter, r *http.Request, err error) {
		if fn := apperr.Override(err); fn != nil {
			fn(w, r)
			return
		}

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
