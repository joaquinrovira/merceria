package v0

import (
	"context"
	"fmt"
	"log"
	"merceria/internal/model"
	"merceria/internal/util"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"google.golang.org/api/sheets/v4"
)

type Operator struct {
	instance atomic.Pointer[Instance]
}

type Instance struct {
	ctx           context.Context
	SpreadsheetId string
	Spreadsheets  *sheets.SpreadsheetsService
	SheetId       int64
	TableId       string
	TableHeader   sheets.GridRange
	Mode          InsertMode
}

type InsertMode int

const (
	ModeUnknown InsertMode = iota
	ModeAppend
	ModePrepend
)

var PrependNames []string = func() []string {
	envNames := os.Getenv("PREPEND_NAMES")
	if os.Getenv("PREPEND_NAMES") == "" {
		return []string{"pedidos", "orders", "index"}
	}

	return util.Map(
		strings.Split(envNames, ","),
		util.TryNormalize,
	)
}()

func New(ctx context.Context, Spreadsheets *sheets.SpreadsheetsService, SpreadsheetId string) (*Operator, error) {
	meta, err := Spreadsheets.Get(SpreadsheetId).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("retrieving spreadsheet metadata: %w", err)
	}

	sheetsmeta := meta.Sheets
	if len(meta.Sheets) < 1 {
		return nil, nil
	}

	operator := &Instance{
		ctx:           ctx,
		Spreadsheets:  Spreadsheets,
		SpreadsheetId: SpreadsheetId,
		Mode:          ModeAppend,
	}

	target := sheetsmeta[0]
	for _, sheet := range sheetsmeta {
		if len(sheet.Tables) != 1 {
			continue
		}
		target = sheet
		table := sheet.Tables[0]
		if slices.Contains(PrependNames, util.Must(util.Normalize(table.Name))) {
			target = sheet
			operator.Mode = ModePrepend
			r := *table.Range
			r.StartRowIndex = r.StartRowIndex + 1
			r.EndRowIndex = r.StartRowIndex
			operator.TableHeader = r
			break
		}

	}

	operator.SheetId = target.Properties.SheetId
	operator.TableId = target.Tables[0].TableId

	o := &Operator{}
	o.instance.Store(operator)

	// Every so often, refresh the operator's metadata to adapt to changes in the spreadsheet structure
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf("Refreshing operator metadata for spreadsheet %s", SpreadsheetId)
				newOperator, err := New(ctx, Spreadsheets, SpreadsheetId)
				if err != nil {
					log.Printf("Error refreshing operator metadata: %v", err)
					continue
				}
				if newOperator == nil {
					log.Printf("Spreadsheet %s no longer has valid structure for operator", SpreadsheetId)
					continue
				}
				o.instance.Store(newOperator.instance.Load())
				log.Printf("Successfully refreshed operator metadata for spreadsheet %s", SpreadsheetId)
				return
			}
		}
	}()

	return o, nil
}

func (o *Operator) Insert(ctx context.Context, rows []model.Row) error {
	return o.instance.Load().Insert(ctx, rows)
}

func (i *Instance) Insert(ctx context.Context, rows []model.Row) error {
	if len(rows) == 0 {
		return nil
	}
	switch i.Mode {
	case ModeAppend:
		return i.append(ctx, rows)
	case ModePrepend:
		return i.prepend(ctx, rows)
	}
	panic(fmt.Errorf("unhandled insert mode: %d", i.Mode))
}

func (i *Instance) append(ctx context.Context, rows []model.Row) error {
	data := util.Map(rows, ToSpreadsheetValuesLegacy)
	_, err := i.Spreadsheets.BatchUpdate(i.SpreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			AppendCells: &sheets.AppendCellsRequest{
				TableId: i.TableId,
				SheetId: i.SheetId,
				Rows:    data,
				Fields:  "*",
			},
		}},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("appending rows: %w", err)
	}
	return nil
}

func (i *Instance) prepend(ctx context.Context, rows []model.Row) error {
	data := util.Map(rows, ToSpreadsheetValues2)

	update := &sheets.BatchUpdateSpreadsheetRequest{Requests: []*sheets.Request{}}

	// 1. Define the request to insert one row below the header
	update.Requests = append(update.Requests, &sheets.Request{
		InsertDimension: &sheets.InsertDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    i.SheetId,
				Dimension:  "ROWS",
				StartIndex: i.TableHeader.StartRowIndex,
				EndIndex:   i.TableHeader.EndRowIndex + int64(len(rows)),
			},
			InheritFromBefore: true,
		},
	})

	// 2. Define the request to update the newly created row with your data
	update.Requests = append(update.Requests, &sheets.Request{
		UpdateCells: &sheets.UpdateCellsRequest{
			Rows:   data,
			Fields: "*",
			Range: &sheets.GridRange{
				SheetId:          i.SheetId,
				StartRowIndex:    i.TableHeader.StartRowIndex,
				EndRowIndex:      i.TableHeader.StartRowIndex + int64(len(rows)),
				StartColumnIndex: i.TableHeader.StartColumnIndex,
				EndColumnIndex:   i.TableHeader.EndColumnIndex,
			},
		},
	})

	_, err := i.Spreadsheets.BatchUpdate(i.SpreadsheetId, update).Context(ctx).Do()
	if err != nil {
		log.Fatalf("prepending rows: %v", err)
	}

	return nil
}
