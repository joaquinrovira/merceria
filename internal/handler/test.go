package handler

import (
	"fmt"
	"merceria/internal/model/apperr"
	"net/http"
)

func Test() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return apperr.WithPublic(fmt.Errorf("this is a test"))
	}
}
