package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"merceria/internal/auth"
	"merceria/internal/middleware"
	"merceria/internal/model"
	"merceria/internal/spreadsheets"
	"merceria/internal/util"
)

func extractIndexAndField(key string) (indexStr, fieldName string, ok bool) {
	if !strings.HasPrefix(key, "data[") {
		return
	}
	bracketAt := strings.Index(key, "]@")
	if bracketAt == -1 {
		return
	}
	indexStr = key[5:bracketAt]
	fieldName = key[bracketAt+2:]
	if indexStr == "" || fieldName == "" {
		return "", "", false
	}
	ok = true
	return
}

func parseForm(r *http.Request) ([]model.Row, error) {
	entries := map[string]map[string]string{}
	hasArrayMode := false

	for key, values := range r.Form {
		if idx, field, ok := extractIndexAndField(key); ok {
			hasArrayMode = true
			if entries[idx] == nil {
				entries[idx] = map[string]string{}
			}
			entries[idx][field] = values[0]
		}
	}

	if hasArrayMode {
		return parseArrayMode(entries)
	}
	return parseSingleMode(r)
}

func parseArrayMode(entries map[string]map[string]string) ([]model.Row, error) {
	indices := make([]string, 0, len(entries))
	for idx := range entries {
		indices = append(indices, idx)
	}
	sortStrings(indices)

	rows := make([]model.Row, 0, len(indices))
	for _, idx := range indices {
		fields := entries[idx]
		row, err := buildRowFromFields(fields, idx)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseSingleMode(r *http.Request) ([]model.Row, error) {
	fields := map[string]string{
		"fecha":    r.FormValue("fecha"),
		"nombre":   r.FormValue("nombre"),
		"telefono": r.FormValue("telefono"),
		"prendas":  r.FormValue("prendas"),
		"notas":    r.FormValue("notas"),
		"grupo":    r.FormValue("grupo"),
		"estado":   r.FormValue("estado"),
	}
	row, err := buildRowFromFields(fields, "")
	if err != nil {
		return nil, err
	}
	return []model.Row{row}, nil
}

func buildRowFromFields(fields map[string]string, indexInfo string) (model.Row, error) {
	var row model.Row

	nombre := fields["nombre"]
	fecha := fields["fecha"]

	if nombre == "" || fecha == "" {
		msg := "Nombre y Fecha son campos requeridos"
		if indexInfo != "" {
			msg += fmt.Sprintf(" at index %s", indexInfo)
		}
		return row, fmt.Errorf("%s", msg)
	}

	if fecha != "" {
		parsed, err := time.Parse("2006-01-02", fecha)
		if err != nil {
			msg := "Formato de fecha inválido, use YYYY-MM-DD"
			if indexInfo != "" {
				msg += fmt.Sprintf(" at index %s", indexInfo)
			}
			return row, fmt.Errorf("%s", msg)
		}
		row.CreatedAt = parsed
	}

	prendas := fields["prendas"]
	if prendas != "" {
		amount, err := strconv.Atoi(prendas)
		if err != nil {
			msg := "Prendas debe ser un número"
			if indexInfo != "" {
				msg += fmt.Sprintf(" at index %s", indexInfo)
			}
			return row, fmt.Errorf("%s", msg)
		}
		row.Amount = amount
	}

	grupo := fields["grupo"]
	validTags := map[string]bool{
		"Grupo 1": true,
		"Grupo 2": true,
		"Grupo 3": true,
	}
	if !validTags[grupo] {
		msg := "Grupo inválido, debe ser 'Grupo 1', 'Grupo 2' o 'Grupo 3'"
		if indexInfo != "" {
			msg += fmt.Sprintf(" at index %s", indexInfo)
		}
		return row, fmt.Errorf("%s", msg)
	}

	estado := fields["estado"]
	if estado != "TRUE" && estado != "FALSE" {
		msg := "Estado inválido, debe ser 'TRUE' o 'FALSE'"
		if indexInfo != "" {
			msg += fmt.Sprintf(" at index %s", indexInfo)
		}
		return row, fmt.Errorf("%s", msg)
	}

	b := make([]byte, 1)
	rand.Read(b)
	orderId := row.CreatedAt.Format("20060102-") + fmt.Sprintf("%X", b)

	row.OrderId = orderId
	row.Name = nombre
	row.Phone = fields["telefono"]
	row.Notes = fields["notas"]
	row.Tag = grupo
	row.Status = estado == "TRUE"

	return row, nil
}

func sortStrings(arr []string) {
	for i := 0; i < len(arr)-1; i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] > arr[j] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
}

func CreateRowForm(ctx context.Context, fs *os.Root) http.HandlerFunc {
	const name = "form.html"
	data := util.ReloadingFile(ctx, fs, name)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(util.Must(data()))
	}
}

func CreateRow(svc *spreadsheets.Service) http.HandlerFunc {

	// TODO: Deduplicate requests by FORM key
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

		rows, err := parseForm(r)
		if err != nil {
			FormError(err.Error())
			return
		}

		ctx := r.Context()
		err = operator.Insert(ctx, rows)
		if err != nil {
			fmt.Printf("failed to insert row: %v\n", err)
			http.Error(w, "failed to create row", http.StatusInternalServerError)
			return
		}

		location := r.URL.RequestURI()
		http.Redirect(w, r, location+"?success=true", http.StatusFound)
	}

}
