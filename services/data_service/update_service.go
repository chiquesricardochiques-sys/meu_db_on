package data_service

import (
	"fmt"
	"strings"

	"meu-provedor/config"
	"meu-provedor/engine/query"
	"meu-provedor/handlers"
)

// ExecuteUpdate monta e executa a query de UPDATE
func ExecuteUpdate(req handlers.UpdateRequest) (int, error) {
	// Obter projectCode
	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, fmt.Errorf("project not found")
	}

	// Montar nome físico da tabela
	table := fmt.Sprintf("%s_%s", projectCode, req.Table)

	// Criar builder
	builder := query.NewUpdate(table)

	// Adicionar campos a atualizar
	for col, val := range req.Data {
		if !query.IsValidIdentifier(col) {
			return 0, fmt.Errorf("invalid column: %s", col)
		}
		builder.Set(col, val)
	}

	// Filtro obrigatório id_instancia
	builder.Where("id_instancia = ?", req.InstanceID)

	// Adicionar filtros simples
	for col, val := range req.Where {
		if !query.IsValidIdentifier(col) {
			return 0, fmt.Errorf("invalid where column: %s", col)
		}
		builder.Where(col+" = ?", val)
	}

	// Filtro raw opcional
	if req.WhereRaw != "" {
		builder.WhereRaw(req.WhereRaw)
	}

	// Gerar query final
	queryStr, args := builder.Build()

	// Executar update
	res, err := config.MasterDB.Exec(queryStr, args...)
	if err != nil {
		return 0, err
	}

	// Retornar quantidade de linhas afetadas
	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}
