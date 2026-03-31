package v0

import (
	"errors"
	"fmt"
	"merceria/internal/model"
	"reflect"
	"strconv"
	"time"
)

// Best effort conversion from an arbitrary row of data into the expected format.
// An error implies some or all fields are missing. The value returned will contained
// as much data as possible.
func FromSpreadsheetData(r []interface{}) (model.Row, error) {
	row := model.Row{}
	var errs []error
	var col = -1

	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		row.OrderId = data
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		t, err := time.Parse("02/01/2006", data) // DD/MM/YYYY
		if err != nil {
			errs = append(errs, fmt.Errorf("column %d: invalid: %w", col, err))
			return
		}
		row.CreatedAt = t
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		row.Name = data
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		row.Phone = data
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		n, err := strconv.ParseInt(data, 10, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("column %d: parsing integer: %w", col, err))
			return
		}
		row.Amount = int(n)
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		row.Notes = data
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		row.Tag = data
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		row.Delayed = data
	}()
	func() {
		col++
		if len(r) <= col {
			errs = append(errs, fmt.Errorf("column %d: missing", col))
			return
		}
		data, ok := r[col].(string)
		if !ok {
			errs = append(errs, fmt.Errorf("column %d: unexpected type: %s", col, reflect.TypeOf(r[col])))
			return
		}
		row.Status = data == "TRUE"
	}()

	return row, errors.Join(errs...)
}
