package query

import (
	"fmt"
	"strings"
)

type InsertBuilder struct {
	Table    string
	Columns  []string
	Values   [][]interface{}
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

func (b *InsertBuilder) Build() (string, []interface{}) {
	placeholders := "(" + strings.Repeat("?,", len(b.Columns)-1) + "?)"
	var allPlaceholders []string
	var allValues []interface{}

	for _, row := range b.Values {
		allPlaceholders = append(allPlaceholders, placeholders)
		allValues = append(allValues, row...)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		b.Table,
		strings.Join(b.Columns, ","),
		strings.Join(allPlaceholders, ","),
	)
	return query, allValues
}
