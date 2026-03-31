package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"merceria/internal/auth"
	"merceria/internal/middleware"
	"merceria/internal/model"
	"merceria/internal/spreadsheets"
	"merceria/internal/util"
)

var names = []string{"Alice", "Bob", "Charlie", "Diana", "Eve"}
var surnames = []string{"Smith", "Johnson", "Williams", "Brown", "Jones"}

func randomName() string {
	return fmt.Sprintf("%s %s", names[rand.Intn(len(names))], surnames[rand.Intn(len(surnames))])
}

func SampleRow() model.Row {
	createdAt := time.Date(2026, time.Month(rand.Intn(12)+1), rand.Intn(30)+1, 0, 0, 0, 0, time.UTC)
	orderId := fmt.Sprintf("%s%03d", createdAt.Format("20060102"), rand.Intn(1e3))
	return model.Row{
		OrderId:   orderId,
		CreatedAt: createdAt,
		Name:      randomName(),
		Phone:     fmt.Sprintf("6%08d", rand.Intn(1e8)),
		Amount:    rand.Intn(49) + 1,
		Notes:     "Example note",
		Tag:       string([...]model.Tag{model.TagGrupo1, model.TagGrupo2, model.TagGrupo3}[rand.Intn(3)]),
		Status:    rand.Intn(2) == 0,
	}
}

func SampleRows(n int) []model.Row {
	rows := make([]model.Row, n)
	for i := range rows {
		rows[i] = SampleRow()
	}
	return rows
}

func InsertRandomRows(svc *spreadsheets.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := ctx.Value(middleware.JwtClaims).(*auth.SessionClaims)
		operator := util.Must(svc.GetOperator(claims.SpreadsheetId))

		nStr := r.PathValue("N")
		n, err := strconv.Atoi(nStr)
		if err != nil || n < 1 {
			http.Error(w, "invalid N; must be a positive integer", http.StatusBadRequest)
			return
		}

		rows := SampleRows(n)
		err = operator.Insert(ctx, rows)
		if err != nil {
			log.Printf("ERROR: appending random rows: %v", err)
			http.Error(w, "failed to append rows", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"rows_inserted": n,
			"data":          rows,
		})
	}
}
