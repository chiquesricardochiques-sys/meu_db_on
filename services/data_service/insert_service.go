package services

import (
	"fmt"
	"sort"
	"strings"
	"meu-provedor/config"
	"meu-provedor/engine/query"
	"meu-provedor/models"
)

// ExecuteInsert executa um INSERT único
func ExecuteInsert(req models.InsertRequest) (int64, error) {
	// Validar requisição
	if req.ProjectID <= 0 {
		return 0, models.ErrInvalidProjectID
	}
	if req.InstanceID <= 0 {
		return 0, models.ErrInvalidInstanceID
	}
	if req.Table == "" {
		return 0, models.ErrTableRequired
	}
	if len(req.Data) == 0 {
		return 0, models.ErrNoDataProvided
	}

	// Obter código do projeto
	projectCode, err := GetProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, err
	}

	// Construir nome da tabela
	table, err := BuildTableName(projectCode, req.Table)
	if err != nil {
		return 0, err
	}

	// Adicionar id_instancia aos dados
	req.Data["id_instancia"] = req.InstanceID

	// ✅ FIX: Extrair colunas em ordem FIXA (sorted)
	var cols []string
	for col := range req.Data {
		if !query.IsValidColumnName(col) {
			return 0, fmt.Errorf("%w: %s", models.ErrInvalidColumn, col)
		}
		cols = append(cols, col)
	}
	
	// ✅ CRÍTICO: Ordenar colunas alfabeticamente
	sort.Strings(cols)

	// ✅ Extrair valores NA MESMA ORDEM das colunas
	var vals []interface{}
	for _, col := range cols {
		vals = append(vals, req.Data[col])
	}

	// Criar InsertBuilder
	builder := query.NewInsert(table, cols)
	builder.AddRow(vals)

	// Executar INSERT
	sqlQuery, args := builder.Build()
	
	// DEBUG (remover em produção)
	fmt.Printf("SQL: %s\nArgs: %v\n", sqlQuery, args)
	
	result, err := config.MasterDB.Exec(sqlQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", models.ErrInsertFailed, err.Error())
	}

	// Retornar ID inserido
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

// ExecuteBatchInsert executa múltiplos INSERTs em lote
func ExecuteBatchInsert(req models.BatchInsertRequest) (int, error) {
	// Validar requisição
	if err := req.Validate(); err != nil {
		return 0, err
	}

	// Obter código do projeto
	projectCode, err := GetProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, err
	}

	// Construir nome da tabela
	table, err := BuildTableName(projectCode, req.Table)
	if err != nil {
		return 0, err
	}

	// ✅ Coletar todas as colunas únicas
	colsMap := make(map[string]bool)
	for _, row := range req.Data {
		for k := range row {
			colsMap[k] = true
		}
	}
	colsMap["id_instancia"] = true

	// ✅ Converter para slice e ORDENAR
	var cols []string
	for col := range colsMap {
		if !query.IsValidColumnName(col) {
			return 0, fmt.Errorf("%w: %s", models.ErrInvalidColumn, col)
		}
		cols = append(cols, col)
	}
	sort.Strings(cols) // ✅ CRÍTICO: ordem fixa

	// ✅ Construir query com placeholders corretos
	var valuePlaceholders []string
	var allValues []interface{}
	
	placeholderCount := 0
	for _, row := range req.Data {
		row["id_instancia"] = req.InstanceID
		
		// Gerar placeholders para esta linha
		var rowPlaceholders []string
		for i := 0; i < len(cols); i++ {
			placeholderCount++
			rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", placeholderCount))
		}
		valuePlaceholders = append(valuePlaceholders, "("+strings.Join(rowPlaceholders, ",")+")")
		
		// Adicionar valores NA ORDEM das colunas
		for _, col := range cols {
			allValues = append(allValues, row[col])
		}
	}

	queryStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(cols, ","),
		strings.Join(valuePlaceholders, ","),
	)

	// DEBUG (remover em produção)
	fmt.Printf("SQL: %s\nArgs: %v\n", queryStr, allValues)

	// Executar batch insert
	_, err = config.MasterDB.Exec(queryStr, allValues...)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", models.ErrInsertFailed, err)
	}

	return len(req.Data), nil
}
