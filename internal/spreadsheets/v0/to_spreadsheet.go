package v0

import (
	"fmt"
	"merceria/internal/model"
	"merceria/internal/util"
	"strconv"

	"google.golang.org/api/sheets/v4"
)

func GridPoint(col, row int64) (string, string) {
	return string(rune('A' + col)), strconv.FormatInt(row+1, 10)
}

func RowToSpreadsheetValuesConverter(grid sheets.GridRange) func([]model.Row) []*sheets.RowData {
	return func(rows []model.Row) []*sheets.RowData {
		result := make([]*sheets.RowData, len(rows))
		for i, row := range rows {
			grid := grid
			grid.StartRowIndex = grid.StartRowIndex + int64(i)
			grid.EndRowIndex = grid.StartRowIndex + 1
			result[i] = ToSpreadsheetValues(row, grid)
		}
		return result
	}
}

const CreatedAtFormat = "2006/01/02"

func ToSpreadsheetValues(r model.Row, grid sheets.GridRange) *sheets.RowData {
	dollarRow := func(col int) string {
		colStr, rowStr := GridPoint(grid.StartColumnIndex+int64(col), grid.StartRowIndex)
		return fmt.Sprintf("$%s%s", colStr, rowStr)
	}
	return &sheets.RowData{
		Values: []*sheets.CellData{
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(fmt.Sprintf("%v", util.Must(strconv.Atoi(r.OrderId)))),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.CreatedAt.Format(CreatedAtFormat)),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.Name),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(fmt.Sprintf("%v", util.Must(strconv.Atoi(r.Phone)))),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					NumberValue: new(float64(r.Amount)),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.Notes),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.Tag),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					FormulaValue: new(fmt.Sprintf(`=IF(OR(%[1]s="";%[2]s="");"";IF(TODAY()-INT(%[1]s)>CHOOSE(VALUE(RIGHT(%[2]s;1));7;10;15);"!!";""))`, dollarRow(1), dollarRow(6))),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					BoolValue: new(r.Status),
				},
			},
		},
	}
}

func ToSpreadsheetValues2(r model.Row) *sheets.RowData {
	realtive := func(src, dst int) string {
		return fmt.Sprintf("INDIRECT(\"RC[%d]\"; FALSE)", dst-src)
	}
	return &sheets.RowData{
		Values: []*sheets.CellData{
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.OrderId),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.CreatedAt.Format(CreatedAtFormat)),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.Name),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.Phone),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					NumberValue: new(float64(r.Amount)),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.Notes),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					StringValue: new(r.Tag),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					FormulaValue: new(fmt.Sprintf(`=IF(OR(%[1]s="";%[2]s="");"";IF(TODAY()-INT(%[1]s)>CHOOSE(VALUE(RIGHT(%[2]s;1));7;10;15);"!!";""))`, realtive(7, 1), realtive(7, 6))),
				},
			},
			{
				UserEnteredValue: &sheets.ExtendedValue{
					BoolValue: new(r.Status),
				},
			},
		},
	}
}

func ToSpreadsheetValuesLegacy(r model.Row) *sheets.RowData {
	realtive := func(src, dst int) string {
		return fmt.Sprintf("INDIRECT(\"RC[%d]\"; FALSE)", dst-src)
	}
	row := ToSpreadsheetValues2(r)
	row.Values[0] = &sheets.CellData{
		UserEnteredValue: &sheets.ExtendedValue{
			FormulaValue: new(fmt.Sprintf(`=IF(%[1]s=""; ""; VALUE(TEXT(INT(%[1]s); "yyyymmdd") & TEXT(COUNTIF(INDIRECT("R1C[1]:RC[1]"; FALSE); INT(%[1]s)); "000")))`, realtive(0, 1))),
		},
	}
	return row
}
