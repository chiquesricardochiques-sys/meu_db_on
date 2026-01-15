package query

import (
	"fmt"
	"strings"
)

type JoinConfig struct {
	Type    string
	Table   string
	Alias   string
	On      string
	Columns []string
}

type JoinSelectBuilder struct {
	BaseTable  string
	BaseAlias  string
	Columns    []string
	Joins      []JoinConfig
	Where      []string
	RawWhere   []string
	GroupBy    string
	Having     string
	OrderBy    string
	Limit      int
	Offset     int
	Values     []interface{}
}

func NewJoinSelect(table, alias string) *JoinSelectBuilder {
	if alias == "" {
		alias = table
	}
	return &JoinSelectBuilder{
		BaseTable: table,
		BaseAlias: alias,
		Columns:   []string{},
	}
}

func (b *JoinSelectBuilder) AddColumns(cols ...string) {
	b.Columns = append(b.Columns, cols...)
}

func (b *JoinSelectBuilder) AddJoin(j JoinConfig) {
	if j.Type == "" {
		j.Type = "INNER"
	}
	b.Joins = append(b.Joins, j)
}

func (b *JoinSelectBuilder) AddWhere(cond string, args ...interface{}) {
	b.Where = append(b.Where, cond)
	b.Values = append(b.Values, args...)
}

func (b *JoinSelectBuilder) AddRawWhere(cond string) {
	b.RawWhere = append(b.RawWhere, cond)
}

func (b *JoinSelectBuilder) Build() (string, []interface{}) {
	if len(b.Columns) == 0 {
		b.Columns = append(b.Columns, "*")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s AS %s",
		strings.Join(b.Columns, ", "),
		b.BaseTable,
		b.BaseAlias,
	)

	for _, j := range b.Joins {
		query += fmt.Sprintf(
			" %s JOIN %s AS %s ON %s",
			strings.ToUpper(j.Type),
			j.Table,
			j.Alias,
			j.On,
		)
	}

	where := []string{}
	where = append(where, b.Where...)
	where = append(where, b.RawWhere...)

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	if b.GroupBy != "" {
		query += " GROUP BY " + b.GroupBy
	}

	if b.Having != "" {
		query += " HAVING " + b.Having
	}

	if b.OrderBy != "" {
		query += " ORDER BY " + b.OrderBy
	}

	if b.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", b.Limit)
		if b.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", b.Offset)
		}
	}

	return query, b.Values
}
