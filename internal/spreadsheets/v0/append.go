package v0

import (
	"context"
	"fmt"
	"merceria/internal/model"
	"merceria/internal/spreadsheets/sheetutil"

	"google.golang.org/api/sheets/v4"
)

func AppendRows(ctx context.Context, Spreadsheets *sheets.SpreadsheetsService, SpreadsheetId string, rows []model.Row) error {
	if len(rows) == 0 {
		return fmt.Errorf("no rows to append")
	}

	meta, err := Spreadsheets.Get(SpreadsheetId).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving spreadsheet metadata: %w", err)
	}

	sheetsmeta := meta.Sheets
	// TODO: Configurable LTR or RTL reading order
	latest := sheetsmeta[len(sheetsmeta)-1]

	tablerange := *latest.Tables[0].Range
	title := latest.Properties.Title

	// Select the really last row with data in the sheet by detecting the last row with data in the first second third column, and appending after that
	values, err := Spreadsheets.Values.Get(SpreadsheetId, sheetutil.AreaSelector(title, tablerange)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("retrieving spreadsheet data: %w", err)
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

	insertAt := tablerange
	insertAt.StartRowIndex = tablerange.StartRowIndex + lastRowWithData + 1
	insertAt.EndRowIndex = insertAt.StartRowIndex + int64(len(rows))

	batch := &sheets.BatchUpdateSpreadsheetRequest{}

	batch.Requests = append(batch.Requests, &sheets.Request{
		InsertDimension: &sheets.InsertDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    latest.Properties.SheetId,
				Dimension:  "ROWS",
				StartIndex: insertAt.StartRowIndex,
				EndIndex:   insertAt.EndRowIndex,
			},
			InheritFromBefore: true,
		},
	})

	data := RowToSpreadsheetValuesConverter(insertAt)(rows)

	batch.Requests = append(batch.Requests, &sheets.Request{
		UpdateCells: &sheets.UpdateCellsRequest{
			Start: &sheets.GridCoordinate{
				SheetId:     latest.Properties.SheetId,
				RowIndex:    insertAt.StartRowIndex,
				ColumnIndex: insertAt.StartColumnIndex,
			},
			Rows:   data,
			Fields: "*",
		},
	})

	_, err = Spreadsheets.BatchUpdate(SpreadsheetId, batch).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("extending pivot table range: %w", err)
	}

	return nil
}
