package query

import (
	"fmt"
	"strings"
	"time"
)

type SoftDeleteBuilder struct {
	Table     string
	Where     []string
	RawWhere  []string
	Values    []interface{}
}

func NewSoftDelete(table string) *SoftDeleteBuilder {
	return &SoftDeleteBuilder{
		Table:  table,
		Where: []string{},
	}
}

func (d *SoftDeleteBuilder) AddWhere(condition string, args ...interface{}) {
	d.Where = append(d.Where, condition)
	d.Values = append(d.Values, args...)
}

func (d *SoftDeleteBuilder) AddRawWhere(condition string) {
	d.RawWhere = append(d.RawWhere, condition)
}

func (d *SoftDeleteBuilder) Build(deletedAt time.Time) (string, []interface{}) {
	where := []string{}
	where = append(where, d.Where...)
	where = append(where, d.RawWhere...)

	query := fmt.Sprintf(
		"UPDATE %s SET deleted_at = ?",
		d.Table,
	)

	args := []interface{}{deletedAt}
	args = append(args, d.Values...)

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	return query, args
}
