package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

func Recover() func(http.Handler) http.Handler {
	defaultLogger := func(format string, args ...any) {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(os.Stderr, format, args...)
	}
	return RecoverWithLogger(defaultLogger)
}
func RecoverWithLogger(logger func(format string, args ...any)) func(http.Handler) http.Handler {
	if logger == nil {
		logger = func(format string, args ...any) {}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger("panic: %v: "+string(debug.Stack()), err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("unhandled panic"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
