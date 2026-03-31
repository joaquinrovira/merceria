package sheetutil

import (
	"merceria/internal/util"
	"strings"
)

func MatchHeaders(row []interface{}, names []string) (hits int) {
	for col := range min(len(row), len(names)) {
		expected := names[col]
		data := row[col]
		actual, ok := data.(string)
		if !ok {
			continue
		}

		cleaned := strings.ToLower(util.Must(util.Normalize(actual)))
		if strings.Contains(cleaned, expected) {
			hits++
		}
	}
	return hits
}
