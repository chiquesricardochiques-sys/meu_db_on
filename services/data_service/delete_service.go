package data_service

import (
	"fmt"
	"meu-provedor/config"
	"meu-provedor/engine/query"
	"meu-provedor/handlers"
)

// ExecuteDelete monta e executa o DELETE
func ExecuteDelete(req handlers.DeleteRequest) (int, error) {
	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, fmt.Errorf("project not found")
	}

	table := fmt.Sprintf("%s_%s", projectCode, req.Table)

	builder := query.NewDelete(table)

	// Filtro obrigatório id_instancia
	builder.Where("id_instancia = ?", req.InstanceID)

	// Filtros simples
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

	// Montar query final
	queryStr, args := builder.Build()

	// Executar DELETE
	res, err := config.MasterDB.Exec(queryStr, args...)
	if err != nil {
		return 0, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}
