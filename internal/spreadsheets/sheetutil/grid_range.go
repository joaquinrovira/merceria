package sheetutil

import (
	"fmt"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func GridRangeToString(g sheets.GridRange) string {
	var b strings.Builder
	b.WriteRune('A' + rune(g.StartColumnIndex))
	b.WriteString(fmt.Sprintf("%d:", g.StartRowIndex+1))
	b.WriteRune('A' + rune(g.EndColumnIndex))
	b.WriteString(fmt.Sprintf("%d", g.EndRowIndex+1))
	return b.String()
}

func AreaSelector(s string, g sheets.GridRange) string {
	return s + "!" + GridRangeToString(g)
}
