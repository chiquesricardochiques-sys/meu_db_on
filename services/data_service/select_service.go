package services

import (
	"fmt"
	"meu-provedor/config"
	"meu-provedor/engine/query"
	"meu-provedor/models"
)

// ============================================================================
// SELECT SERVICE
// ============================================================================

// ExecuteAdvancedSelect executa um SELECT avançado com suporte a JOINs
func ExecuteAdvancedSelect(req models.AdvancedSelectRequest) ([]map[string]interface{}, error) {
	// Validar requisição
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Obter código do projeto
	projectCode, err := GetProjectCodeByID(req.ProjectID)
	if err != nil {
		return nil, err
	}

	// Construir nome físico da tabela
	mainTable, err := BuildTableName(projectCode, req.Table)
	if err != nil {
		return nil, err
	}

	mainAlias := req.Alias
	if mainAlias == "" {
		mainAlias = mainTable
	}

	// Criar SelectBuilder
	builder := query.NewSelect(mainTable, mainAlias)

	// Definir colunas
	if len(req.Select) > 0 {
		builder.SetColumns(req.Select)
	}

	// Adicionar JOINs
	for _, j := range req.Joins {
		joinTable, err := BuildTableName(projectCode, j.Table)
		if err != nil {
			return nil, err
		}
		builder.AddJoin(j.Type, joinTable, j.Alias, j.On)
	}

	// Filtro obrigatório: id_instancia
	builder.AddWhere(fmt.Sprintf("%s.id_instancia = ?", mainAlias), req.InstanceID)

	// Filtros simples (WHERE)
	for k, v := range req.Where {
		if !query.IsValidColumnName(k) {
			return nil, fmt.Errorf("%w: %s", models.ErrInvalidColumn, k)
		}
		builder.AddWhere(k+" = ?", v)
	}

	// Filtro raw (WHERE customizado)
	if req.WhereRaw != "" {
		builder.AddWhere("(" + req.WhereRaw + ")")
	}

	// GROUP BY
	if req.GroupBy != "" {
		builder.SetGroupBy(req.GroupBy)
	}

	// HAVING
	if req.Having != "" {
		builder.SetHaving(req.Having)
	}

	// ORDER BY
	if req.OrderBy != "" {
		builder.SetOrderBy(req.OrderBy)
	}

	// LIMIT e OFFSET
	if req.Limit > 0 {
		builder.SetLimitOffset(req.Limit, req.Offset)
	}

	// Executar query
	rows, err := config.MasterDB.Query(builder.Build(), builder.GetValues()...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrQueryFailed, err)
	}
	defer rows.Close()

	// Converter para map
	result, err := RowsToMap(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteAdvancedJoinSelect executa SELECT com múltiplos JOINs complexos
func ExecuteAdvancedJoinSelect(req models.AdvancedJoinSelectRequest) ([]map[string]interface{}, error) {
	// Validar requisição básica
	if req.ProjectID <= 0 {
		return nil, models.ErrInvalidProjectID
	}
	if req.InstanceID <= 0 {
		return nil, models.ErrInvalidInstanceID
	}
	if req.Base.Table == "" {
		return nil, models.ErrTableRequired
	}

	// Obter código do projeto
	projectCode, err := GetProjectCodeByID(req.ProjectID)
	if err != nil {
		return nil, err
	}

	// Construir nome da tabela base
	baseTable, err := BuildTableName(projectCode, req.Base.Table)
	if err != nil {
		return nil, err
	}

	// Criar JoinSelectBuilder
	builder := query.NewJoinSelect(baseTable, req.Base.Alias)

	// Adicionar colunas da tabela base
	if len(req.Base.Columns) > 0 {
		builder.AddColumns(req.Base.Columns...)
	}

	// Adicionar JOINs
	for _, j := range req.Joins {
		joinTable, err := BuildTableName(projectCode, j.Table)
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

		// Adicionar colunas do JOIN
		if len(j.Columns) > 0 {
			builder.AddColumns(j.Columns...)
		}
	}

	// Filtro obrigatório: id_instancia na tabela base
	baseAlias := req.Base.Alias
	if baseAlias == "" {
		baseAlias = baseTable
	}
	builder.AddWhere(fmt.Sprintf("%s.id_instancia = ?", baseAlias), req.InstanceID)

	// Filtros simples (WHERE)
	for k, v := range req.Where {
		if !query.IsValidColumnName(k) {
			return nil, fmt.Errorf("%w: %s", models.ErrInvalidColumn, k)
		}
		builder.AddWhere(fmt.Sprintf("%s = ?", k), v)
	}

	// Filtros raw
	for _, raw := range req.WhereRaw {
		builder.AddRawWhere(raw)
	}

	// GROUP BY, HAVING, ORDER BY
	builder.SetGroupBy(req.GroupBy)
	builder.SetHaving(req.Having)
	builder.SetOrderBy(req.OrderBy)

	// LIMIT e OFFSET
	builder.SetLimitOffset(req.Limit, req.Offset)

	// Executar query
	sqlQuery, args := builder.Build()
	rows, err := config.MasterDB.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrQueryFailed, err)
	}
	defer rows.Close()

	// Converter para map
	result, err := RowsToMap(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}