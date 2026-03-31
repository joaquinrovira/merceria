package v0

import (
	"errors"
	"fmt"
	"merceria/internal/model"
	"time"

	"google.golang.org/api/sheets/v4"
)

// Best effort conversion from an arbitrary row of data into the expected format.
// An error implies some or all fields are missing. The value returned will contained
// as much data as possible.
func FromSpreadsheetRaw(r *sheets.RowData) (model.Row, error) {
	row := model.Row{}
	var errs []error
	var col = -1

	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		if r.Values[col].UserEnteredValue.StringValue == nil {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		}
		row.OrderId = *r.Values[col].UserEnteredValue.StringValue
	}()
	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		switch {
		default:
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		case r.Values[col].UserEnteredValue.StringValue != nil:
			t, err := time.Parse(CreatedAtFormat, *r.Values[col].UserEnteredValue.StringValue)
			if err != nil {
				errs = append(errs, fmt.Errorf("column %d: invalid: %w", col, err))
				return
			}
			row.CreatedAt = t
		case r.Values[col].UserEnteredValue.NumberValue != nil:
			days := int(*r.Values[col].UserEnteredValue.NumberValue)
			epoch := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
			t := epoch.AddDate(0, 0, int(days))
			row.CreatedAt = t
		}
	}()
	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		if r.Values[col].UserEnteredValue.StringValue == nil {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		}
		row.Name = *r.Values[col].UserEnteredValue.StringValue
	}()
	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		if r.Values[col].UserEnteredValue.NumberValue == nil {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		}
		row.Phone = fmt.Sprintf("%.0f", *r.Values[col].UserEnteredValue.NumberValue)
	}()
	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		if r.Values[col].UserEnteredValue.NumberValue == nil {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		}
		row.Amount = int(*r.Values[col].UserEnteredValue.NumberValue)
	}()
	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		if r.Values[col].UserEnteredValue.StringValue == nil {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		}
		row.Notes = *r.Values[col].UserEnteredValue.StringValue
	}()
	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		if r.Values[col].UserEnteredValue.StringValue == nil {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		}
		row.Tag = *r.Values[col].UserEnteredValue.StringValue
	}()
	col++ // Skip column: [Retraso]
	func() {
		col++
		if len(r.Values) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		if r.Values[col].UserEnteredValue == nil {
			errs = append(errs, fmt.Errorf("column %d: empty value", col))
			return
		}
		if r.Values[col].UserEnteredValue.BoolValue == nil {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, ExtendedValueType(r.Values[col].UserEnteredValue)))
			return
		}
		row.Status = *r.Values[col].UserEnteredValue.BoolValue
	}()

	return row, errors.Join(errs...)
}

func ExtendedValueType(v *sheets.ExtendedValue) string {
	switch {
	case v.BoolValue != nil:
		return "Bool"
	case v.ErrorValue != nil:
		return "Error"
	case v.FormulaValue != nil:
		return "Formula"
	case v.NumberValue != nil:
		return "Number"
	case v.StringValue != nil:
		return "String"
	}
	return "nil"
}
