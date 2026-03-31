package v0

import (
	"context"
	"fmt"
	"merceria/internal/spreadsheets/sheetutil"
	"slices"

	"google.golang.org/api/sheets/v4"
)

func RetrieveRows(ctx context.Context, Spreadsheets *sheets.SpreadsheetsService, SpreadsheetId string, n int) ([][]interface{}, error) {
	if n <= 0 {
		return nil, nil
	}

	meta, err := Spreadsheets.Get(SpreadsheetId).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("retrieving spreadsheet metadata: %w", err)
	}

	sheetsmeta := meta.Sheets
	latest := sheetsmeta[len(sheetsmeta)-1]

	tablerange := *latest.Tables[0].Range
	title := latest.Properties.Title

	// Select the really last row with data in the sheet by detecting the last row with data in the first second third column, and appending after that
	values, err := Spreadsheets.Values.Get(SpreadsheetId, title+"!"+sheetutil.GridRangeToString(tablerange)).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("retrieving spreadsheet data: %w", err)
	}
	if values.MajorDimension != "ROWS" {
		panic("unexpected major dimension: " + values.MajorDimension)
	}
	lastRowWithData := int64(0)
	for i, row := range values.Values {
		if len(row) > 0 {
			lastRowWithData = int64(i)
		}
	}

	selectFrom := tablerange
	selectFrom.StartRowIndex = tablerange.StartRowIndex + lastRowWithData - int64(n) + 1
	selectFrom.EndRowIndex = selectFrom.StartRowIndex + int64(n)

	data, err := Spreadsheets.Values.Get(SpreadsheetId, title+"!"+sheetutil.GridRangeToString(selectFrom)).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("extending pivot table range: %w", err)
	}
	if values.MajorDimension != "ROWS" {
		panic("unexpected major dimension: " + values.MajorDimension)
	}

	v := data.Values
	slices.Reverse(v)
	return v, nil
}

func RetrieveRowsRaw(ctx context.Context, Spreadsheets *sheets.SpreadsheetsService, SpreadsheetId string, n int) ([]*sheets.RowData, error) {
	if n <= 0 {
		return nil, nil
	}

	meta, err := Spreadsheets.Get(SpreadsheetId).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("retrieving spreadsheet metadata: %w", err)
	}

	sheetsmeta := meta.Sheets
	latest := sheetsmeta[len(sheetsmeta)-1]

	tablerange := *latest.Tables[0].Range
	title := latest.Properties.Title

	// Select the really last row with data in the sheet by detecting the last row with data in the first second third column, and appending after that
	values, err := Spreadsheets.Values.Get(SpreadsheetId, title+"!"+sheetutil.GridRangeToString(tablerange)).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("retrieving spreadsheet data: %w", err)
	}

	if values.MajorDimension != "ROWS" {
		panic("unexpected major dimension: " + values.MajorDimension)
	}
	lastRowWithData := int64(0)
	for i, row := range values.Values {
		if len(row) > 0 {
			lastRowWithData = int64(i)
		}
	}

	selectFrom := tablerange
	selectFrom.StartRowIndex = tablerange.StartRowIndex + lastRowWithData - int64(n) + 1
	selectFrom.EndRowIndex = selectFrom.StartRowIndex + int64(n) - 1

	data, err := Spreadsheets.Get(SpreadsheetId).Ranges(title + "!" + sheetutil.GridRangeToString(selectFrom)).IncludeGridData(true).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("extending pivot table range: %w", err)
	}

	v := data.Sheets[0].Data[0].RowData
	slices.Reverse(v)
	return v, nil
}
