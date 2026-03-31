package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"merceria/internal/auth"
	"merceria/internal/middleware"
	"merceria/internal/model"
	"merceria/internal/spreadsheets"
	"merceria/internal/util"
)

func ShittyFsNotify(ctx context.Context, fs *os.Root, name string) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		prev, _ := fs.Stat(name)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(200 * time.Millisecond):
			}

			next, _ := fs.Stat(name)
			if next.ModTime() != prev.ModTime() {
				prev = next
				ch <- struct{}{}
			}
		}
	}()
	return ch
}

func ReloadingFile(ctx context.Context, fs *os.Root, name string) func() (data []byte, err error) {
	data, err := fs.ReadFile(name)
	go func() {
		for range ShittyFsNotify(ctx, fs, name) {
			log.Printf("%s changed, reloading", name)
			data, err = fs.ReadFile(name)
		}
	}()

	return func() ([]byte, error) { return data, err }
}

func CreateRowForm(ctx context.Context, fs *os.Root) http.HandlerFunc {
	const name = "form.html"
	data := ReloadingFile(ctx, fs, name)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(util.Must(data()))
	}
}

func CreateRow(svc *spreadsheets.Service) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		FormError := func(msg string) {
			http.Redirect(w, r, "/form?error="+url.QueryEscape(msg), http.StatusSeeOther)
		}

		claims := r.Context().Value(middleware.JwtClaims).(*auth.SessionClaims)
		operator := util.Must(svc.GetOperator(claims.SpreadsheetId))

		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form data", http.StatusBadRequest)
			return
		}

		fecha := r.FormValue("fecha")
		nombre := r.FormValue("nombre")
		telefono := r.FormValue("telefono")
		prendas := r.FormValue("prendas")
		notas := r.FormValue("notas")
		grupo := r.FormValue("grupo")
		estado := r.FormValue("estado")

		if nombre == "" || fecha == "" {
			FormError("Nombre y Fecha son campos requeridos")
			return
		}

		var createdAt time.Time
		if fecha != "" {
			parsed, err := time.Parse("2006-01-02", fecha)
			if err != nil {
				FormError("Formato de fecha inválido, use YYYY-MM-DD")
				return
			}
			createdAt = parsed
		}

		amount, err := strconv.Atoi(prendas)
		if err != nil && prendas != "" {
			FormError("Prendas debe ser un número")
			return
		}

		validTags := map[string]bool{
			"Grupo 1": true,
			"Grupo 2": true,
			"Grupo 3": true,
		}
		if !validTags[grupo] {
			FormError("Grupo inválido, debe ser 'Grupo 1', 'Grupo 2' o 'Grupo 3'")
			return
		}

		if estado != "TRUE" && estado != "FALSE" {
			FormError("Estado inválido, debe ser 'TRUE' o 'FALSE'")
			return
		}

		b := make([]byte, 1)
		rand.Read(b)
		orderId := createdAt.Format("20060102-") + fmt.Sprintf("%X", b)

		row := model.Row{
			OrderId:   orderId,
			CreatedAt: createdAt,
			Name:      nombre,
			Phone:     telefono,
			Amount:    amount,
			Notes:     notas,
			Tag:       grupo,
			Status:    estado == "TRUE",
		}

		ctx := r.Context()
		err = operator.Insert(ctx, []model.Row{row})
		if err != nil {
			fmt.Printf("failed to insert row: %v\n", err)
			http.Error(w, "failed to create row", http.StatusInternalServerError)
			return
		}

		location := r.URL.RequestURI()
		http.Redirect(w, r, location+"?success=true", http.StatusFound)
	}

}
