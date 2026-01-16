package data_service

import (
	"fmt"
	"strings"

	"meu-provedor/config"
	"meu-provedor/models"
	"meu-provedor/engine/query"
)

/*
====================================================
EXECUTE SELECT - SERVIÇO
====================================================
*/

func ExecuteSelect(req models.AdvancedQueryRequest) ([]map[string]interface{}, error) {

	// Obter prefixo do projeto
	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}

	// Nome físico da tabela
	mainTable := fmt.Sprintf("%s_%s", projectCode, req.Table)
	mainAlias := req.Alias
	if mainAlias == "" {
		mainAlias = mainTable
	}

	// Criar SelectBuilder
	builder := query.NewSelect(mainTable, mainAlias)

	// Colunas
	if len(req.Select) > 0 {
		builder.SetColumns(req.Select)
	}

	// Joins
	for _, j := range req.Joins {
		builder.AddJoin(j.Type, fmt.Sprintf("%s_%s", projectCode, j.Table), j.Alias, j.On)
	}

	// Where simples
	builder.AddWhere(fmt.Sprintf("%s.id_instancia = ?", mainAlias), req.InstanceID)
	for k, v := range req.Where {
		builder.AddWhere(k+" = ?", v)
	}

	// Where raw
	if req.WhereRaw != "" {
		builder.AddWhere("("+req.WhereRaw+")")
	}

	// Group, Having, Order, Limit
	if req.GroupBy != "" {
		builder.SetGroupBy(req.GroupBy)
	}
	if req.Having != "" {
		builder.Having = req.Having // Campo extra que podemos adicionar no SelectBuilder
	}
	if req.OrderBy != "" {
		builder.SetOrderBy(req.OrderBy)
	}
	if req.Limit > 0 {
		builder.SetLimitOffset(req.Limit, req.Offset)
	}

	// Executar query
	rows, err := config.MasterDB.Query(builder.Build(), builder.Values...)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	// Converter para slice de map
	return rowsToMap(rows), nil
}

/*
====================================================
HELPER - CONVERTER ROWS PARA MAP
====================================================
*/

func rowsToMap(rows *sql.Rows) []map[string]interface{} {
	cols, _ := rows.Columns()
	var result []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		rows.Scan(ptrs...)

		row := make(map[string]interface{})
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result = append(result, row)
	}

	return result
}

/*
====================================================
HELPER - OBTER CÓDIGO DO PROJETO
====================================================
*/

func getProjectCodeByID(projectID int64) (string, error) {
	var code string
	err := config.MasterDB.QueryRow("SELECT code FROM projects WHERE id = ?", projectID).Scan(&code)
	if err != nil {
		return "", err
	}
	return code, nil
}
