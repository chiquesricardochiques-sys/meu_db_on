package query

import (
	"fmt"
	"strings"
)

// ============================================================================
// PLACEHOLDER BUILDER - MySQL
// ============================================================================


// ============================================================================
// INSERT BUILDER
// ============================================================================

type InsertBuilder struct {
	Table   string
	Columns []string
	Values  [][]interface{}
}

func NewInsert(table string, columns []string) *InsertBuilder {
	return &InsertBuilder{
		Table:   table,
		Columns: columns,
		Values:  [][]interface{}{},
	}
}

func (b *InsertBuilder) AddRow(row []interface{}) *InsertBuilder {
	b.Values = append(b.Values, row)
	return b
}

// Build gera query MySQL com placeholders ?
func (b *InsertBuilder) Build() (string, []interface{}) {
	var allPlaceholders []string
	var allValues []interface{}

	for _, row := range b.Values {
		// Criar placeholders: (?, ?, ?)
		placeholders := BuildPlaceholders(len(b.Columns))
		allPlaceholders = append(allPlaceholders, placeholders)
		allValues = append(allValues, row...)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		b.Table,
		strings.Join(b.Columns, ","),
		strings.Join(allPlaceholders, ","),
	)

	return query, allValues
}

// ============================================================================
// VALIDAÇÃO
// ============================================================================

func IsValidColumnName(col string) bool {
	if col == "" || len(col) > 64 {
		return false
	}
	for _, c := range col {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || 
		     (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}


