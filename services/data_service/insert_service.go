package services

import (
	"fmt"
	"strings"
	"meu-provedor/config"
	"meu-provedor/engine/query"
	"meu-provedor/models"
)

// ============================================================================
// INSERT SERVICE
// ============================================================================

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

	// Extrair colunas e valores
	var cols []string
	var vals []interface{}

	for col, val := range req.Data {
		if !query.IsValidColumnName(col) {
			return 0, fmt.Errorf("%w: %s", models.ErrInvalidColumn, col)
		}
		cols = append(cols, col)
		vals = append(vals, val)
	}

	// Criar InsertBuilder
	builder := query.NewInsert(table, cols)
	builder.AddRow(vals)

	// Executar INSERT
	sqlQuery, args := builder.Build()
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

	// Coletar todas as colunas únicas
	colsMap := make(map[string]bool)
	for _, row := range req.Data {
		for k := range row {
			colsMap[k] = true
		}
	}
	colsMap["id_instancia"] = true // obrigatória

	// Converter mapa para slice ordenado
	var cols []string
	for col := range colsMap {
		if !query.IsValidColumnName(col) {
			return 0, fmt.Errorf("%w: %s", models.ErrInvalidColumn, col)
		}
		cols = append(cols, col)
	}

	// Construir query manualmente (batch insert)
	placeholders := query.BuildPlaceholders(len(cols))
	queryStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, strings.Join(cols, ","))

	var allValues []interface{}
	var valuePlaceholders []string

	for _, row := range req.Data {
		row["id_instancia"] = req.InstanceID
		
		var rowValues []interface{}
		for _, col := range cols {
			rowValues = append(rowValues, row[col])
		}
		
		allValues = append(allValues, rowValues...)
		valuePlaceholders = append(valuePlaceholders, placeholders)
	}

	queryStr += strings.Join(valuePlaceholders, ",")

	// Executar batch insert
	_, err = config.MasterDB.Exec(queryStr, allValues...)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", models.ErrInsertFailed, err)
	}

	return len(req.Data), nil

}
