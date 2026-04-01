package middleware

import (
	"net/http"
	"slices"
)

// From composes the given middlewares into a single middleware in left-to-right order.
func From(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for _, fn := range slices.Backward(middlewares) {
			next = fn(next)
		}
		return next
	}
}

func Apply(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return From(middlewares...)(handler)
}

func ApplyFunc(handler http.HandlerFunc, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return Apply(handler, middlewares...)
}
