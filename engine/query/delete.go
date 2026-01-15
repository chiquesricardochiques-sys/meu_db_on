package query

import (
	"fmt"
	"strings"
)

// DeleteBuilder armazena as partes do DELETE
type DeleteBuilder struct {
	Table        string
	WhereClauses []string
	WhereValues  []interface{}
}

// NewDelete cria um builder
func NewDelete(table string) *DeleteBuilder {
	return &DeleteBuilder{
		Table: table,
		WhereClauses: []string{},
		WhereValues: []interface{}{},
	}
}

// Where adiciona condição WHERE
func (d *DeleteBuilder) Where(condition string, args ...interface{}) *DeleteBuilder {
	d.WhereClauses = append(d.WhereClauses, condition)
	d.WhereValues = append(d.WhereValues, args...)
	return d
}

// WhereRaw adiciona filtro customizado
func (d *DeleteBuilder) WhereRaw(raw string) *DeleteBuilder {
	d.WhereClauses = append(d.WhereClauses, "("+raw+")")
	return d
}

// Build gera query final
func (d *DeleteBuilder) Build() (string, []interface{}) {
	wherePart := strings.Join(d.WhereClauses, " AND ")
	return fmt.Sprintf("DELETE FROM %s WHERE %s", d.Table, wherePart), d.WhereValues
}
