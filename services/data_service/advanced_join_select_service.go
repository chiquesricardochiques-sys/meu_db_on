package data_service

import (
	"fmt"

	"meu-provedor/config"
	"meu-provedor/engine/query"
)

/*
====================================================
EXECUTOR – ADVANCED JOIN SELECT
====================================================
*/

func ExecuteAdvancedJoinSelect(req AdvancedJoinSelectRequest) ([]map[string]interface{}, error) {
	// resolve projeto
	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		return nil, err
	}

	// tabela base com prefixo
	baseTable, err := buildTableName(projectCode, req.Base.Table)
	if err != nil {
		return nil, err
	}

	builder := query.NewJoinSelect(baseTable, req.Base.Alias)

	// colunas da tabela base
	if len(req.Base.Columns) > 0 {
		builder.AddColumns(req.Base.Columns...)
	}

	// JOINS
	for _, j := range req.Joins {
		joinTable, err := buildTableName(projectCode, j.Table)
		if err != nil {
			return nil, err
		}

		builder.AddJoin(query.JoinConfig{
			Type:    j.Type,
			Table:   joinTable,
			Alias:   j.Alias,
			On:      j.On,
			Columns: j.Columns,
		})

		if len(j.Columns) > 0 {
			builder.AddColumns(j.Columns...)
		}
	}

	// isolamento por instância (SEMPRE na tabela base)
	baseAlias := req.Base.Alias
	if baseAlias == "" {
		baseAlias = baseTable
	}

	builder.AddWhere(fmt.Sprintf("%s.id_instancia = ?", baseAlias), req.InstanceID)

	// WHERE simples
	for k, v := range req.Where {
		builder.AddWhere(fmt.Sprintf("%s = ?", k), v)
	}

	// WHERE RAW
	for _, raw := range req.WhereRaw {
		builder.AddRawWhere(raw)
	}

	// GROUP / HAVING / ORDER
	builder.GroupBy = req.GroupBy
	builder.Having = req.Having
	builder.OrderBy = req.OrderBy
	builder.Limit = req.Limit
	builder.Offset = req.Offset

	// build final
	sqlQuery, args := builder.Build()

	rows, err := config.MasterDB.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToMap(rows), nil
}
