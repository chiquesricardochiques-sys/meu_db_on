package query

import (
	"fmt"
	"strings"
)

// SelectBuilder armazena as partes do SELECT
type SelectBuilder struct {
	Table     string
	Alias     string
	Columns   []string
	Joins     []string
	Where     []string
	GroupBy   string
	OrderBy   string
	Limit     int
	Offset    int
	Values    []interface{}
}

// NewSelect cria um builder inicial
func NewSelect(table, alias string) *SelectBuilder {
	if alias == "" {
		alias = table
	}
	return &SelectBuilder{
		Table: table,
		Alias: alias,
		Columns: []string{"*"},
		Joins: []string{},
		Where: []string{},
		Values: []interface{}{},
	}
}

// SetColumns define colunas do SELECT
func (s *SelectBuilder) SetColumns(cols []string) *SelectBuilder {
	if len(cols) > 0 {
		s.Columns = cols
	}
	return s
}

// AddJoin adiciona JOIN
func (s *SelectBuilder) AddJoin(joinType, table, alias, on string) *SelectBuilder {
	if joinType == "" {
		joinType = "INNER"
	}
	if alias == "" {
		alias = table
	}
	s.Joins = append(s.Joins, fmt.Sprintf("%s JOIN %s AS %s ON %s", strings.ToUpper(joinType), table, alias, on))
	return s
}

// AddWhere adiciona condição WHERE
func (s *SelectBuilder) AddWhere(condition string, args ...interface{}) *SelectBuilder {
	s.Where = append(s.Where, condition)
	s.Values = append(s.Values, args...)
	return s
}

// SetGroupBy
func (s *SelectBuilder) SetGroupBy(group string) *SelectBuilder {
	s.GroupBy = group
	return s
}

// SetOrderBy
func (s *SelectBuilder) SetOrderBy(order string) *SelectBuilder {
	s.OrderBy = order
	return s
}

// SetLimitOffset
func (s *SelectBuilder) SetLimitOffset(limit, offset int) *SelectBuilder {
	s.Limit = limit
	s.Offset = offset
	return s
}

// Build gera query final
func (s *SelectBuilder) Build() string {
	query := fmt.Sprintf("SELECT %s FROM %s AS %s", strings.Join(s.Columns, ", "), s.Table, s.Alias)
	if len(s.Joins) > 0 {
		query += " " + strings.Join(s.Joins, " ")
	}
	if len(s.Where) > 0 {
		query += " WHERE " + strings.Join(s.Where, " AND ")
	}
	if s.GroupBy != "" {
		query += " GROUP BY " + s.GroupBy
	}
	if s.OrderBy != "" {
		query += " ORDER BY " + s.OrderBy
	}
	if s.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", s.Limit)
		if s.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", s.Offset)
		}
	}
	return query
}
