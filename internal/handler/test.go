package handler

import (
	"fmt"
	apperr "merceria/internal/model/error"
	"net/http"
)

func Test() Handler {
	return func(r *http.Request) (func(w http.ResponseWriter), error) {
		return nil, apperr.WithPublic(fmt.Errorf("this is a test"))
	}
}
