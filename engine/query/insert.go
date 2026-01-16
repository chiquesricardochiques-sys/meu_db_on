package query

import (
	"fmt"
	"strings"
)

// ============================================================================
// INSERT BUILDER
// ============================================================================

// InsertBuilder constrói queries INSERT
type InsertBuilder struct {
	Table   string
	Columns []string
	Values  [][]interface{}
}

// NewInsert cria um novo InsertBuilder
func NewInsert(table string, columns []string) *InsertBuilder {
	return &InsertBuilder{
		Table:   table,
		Columns: columns,
		Values:  [][]interface{}{},
	}
}

// AddRow adiciona uma linha de valores
func (b *InsertBuilder) AddRow(row []interface{}) *InsertBuilder {
	b.Values = append(b.Values, row)
	return b
}

// Build gera a query SQL final e retorna valores achatados
func (b *InsertBuilder) Build() (string, []interface{}) {
	placeholders := BuildPlaceholders(len(b.Columns))
	
	var allPlaceholders []string
	var allValues []interface{}

	for _, row := range b.Values {
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