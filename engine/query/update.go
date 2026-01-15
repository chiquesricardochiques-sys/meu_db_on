package query

import (
	"fmt"
	"strings"
)

// UpdateBuilder armazena as partes do UPDATE
type UpdateBuilder struct {
	Table string
	Sets  []string
	WhereClauses []string
	WhereValues  []interface{}
}

// NewUpdate cria um builder
func NewUpdate(table string) *UpdateBuilder {
	return &UpdateBuilder{
		Table: table,
		Sets: []string{},
		WhereClauses: []string{},
		WhereValues: []interface{}{},
	}
}

// Set adiciona colunas a atualizar
func (u *UpdateBuilder) Set(col string, val interface{}) *UpdateBuilder {
	u.Sets = append(u.Sets, fmt.Sprintf("%s = ?", col))
	u.WhereValues = append(u.WhereValues, val) // temporário, depois ajustamos
	return u
}

// Where adiciona condição WHERE
func (u *UpdateBuilder) Where(condition string, args ...interface{}) *UpdateBuilder {
	u.WhereClauses = append(u.WhereClauses, condition)
	u.WhereValues = append(u.WhereValues, args...)
	return u
}

// WhereRaw adiciona filtro customizado
func (u *UpdateBuilder) WhereRaw(raw string) *UpdateBuilder {
	u.WhereClauses = append(u.WhereClauses, "("+raw+")")
	return u
}

// Build gera a query final
func (u *UpdateBuilder) Build() (string, []interface{}) {
	setPart := strings.Join(u.Sets, ", ")
	wherePart := strings.Join(u.WhereClauses, " AND ")
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", u.Table, setPart, wherePart), u.WhereValues
}

// Função auxiliar de validação de nomes
func IsValidIdentifier(s string) bool {
	// Simples: apenas letras, números e underscore
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || 
			(c >= 'A' && c <= 'Z') || 
			(c >= '0' && c <= '9') || 
			c == '_') {
			return false
		}
	}
	return true
}
