package model

import (
	"time"
)

// [Row] represents a single row in the spreadsheet, with fields corresponding to the columns
type Row struct {
	OrderId   string    // YYYYMMDDnnnn format
	CreatedAt time.Time // Date in YYYY/MM/DD format
	Name      string    // Client name
	Phone     string    // Phone number
	Amount    int       // Item count
	Notes     string    // Arbitrary string
	Tag       string    // Tag: "Grupo 1", "Grupo 2", "Grupo 3"
	Delayed   string    // Calculated by Excel formula
	Status    bool
}

// [Tag] defines the valid values for the [Row.Tag] field
type Tag string

const (
	TagGrupo1 Tag = "Grupo 1"
	TagGrupo2 Tag = "Grupo 2"
	TagGrupo3 Tag = "Grupo 3"
)
