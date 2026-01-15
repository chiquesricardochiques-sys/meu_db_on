package query

import (
	"fmt"
	"strings"
)

type AggregateBuilder struct {
	Table     string
	Alias     string
	Operation string
	Column    string
	Where     []string
	Values    []interface{}
}

func NewAggregate(table, alias, operation, column string) *AggregateBuilder {
	if alias == "" {
		alias = table
	}
	return &AggregateBuilder{
		Table:     table,
		Alias:     alias,
		Operation: strings.ToUpper(operation),
		Column:    column,
		Where:     []string{},
		Values:    []interface{}{},
	}
}

func (a *AggregateBuilder) AddWhere(condition string, args ...interface{}) {
	a.Where = append(a.Where, condition)
	a.Values = append(a.Values, args...)
}

func (a *AggregateBuilder) Build() string {
	target := "*"
	if a.Column != "" {
		target = a.Column
	}

	var selectExpr string
	if a.Operation == "EXISTS" {
		selectExpr = "EXISTS(SELECT 1"
	} else {
		selectExpr = fmt.Sprintf("%s(%s)", a.Operation, target)
	}

	query := fmt.Sprintf("SELECT %s FROM %s AS %s", selectExpr, a.Table, a.Alias)

	if len(a.Where) > 0 {
		query += " WHERE " + strings.Join(a.Where, " AND ")
	}

	if a.Operation == "EXISTS" {
		query += ")"
	}

	return query
}
