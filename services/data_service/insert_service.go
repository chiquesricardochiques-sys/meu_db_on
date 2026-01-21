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

	// ✅ Manter a construção do nome da tabela como estava
	// Exemplo: "salao_beleza" + "_" + "profissionais" = "salao_beleza_profissionais"
	table := fmt.Sprintf("%s_%s", projectCode, req.Table)

	// Adicionar id_instancia aos dados
	req.Data["id_instancia"] = req.InstanceID

	// Extrair colunas em ordem fixa (alfabética)
	var cols []string
	for col := range req.Data {
		if !query.IsValidColumnName(col) {
			return 0, fmt.Errorf("%w: %s", models.ErrInvalidColumn, col)
		}
		cols = append(cols, col)
	}
	sort.Strings(cols)

	// Extrair valores na mesma ordem das colunas
	var vals []interface{}
	for _, col := range cols {
		vals = append(vals, req.Data[col])
	}

	// Criar InsertBuilder
	builder := query.NewInsert(table, cols)
	builder.AddRow(vals)

	// Executar INSERT
	sqlQuery, args := builder.Build()
	
	fmt.Printf("📝 SQL: %s\n", sqlQuery)
	fmt.Printf("📊 Args: %v\n", args)
	
	result, err := config.MasterDB.Exec(sqlQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", models.ErrInsertFailed, err.Error())
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

// ExecuteBatchInsert executa múltiplos INSERTs em lote
func ExecuteBatchInsert(req models.BatchInsertRequest) (int, error) {
	if err := req.Validate(); err != nil {
		return 0, err
	}

	projectCode, err := GetProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, err
	}

	// ✅ Manter a construção do nome da tabela como estava
	table := fmt.Sprintf("%s_%s", projectCode, req.Table)

	// Coletar todas as colunas únicas
	colsMap := make(map[string]bool)
	for _, row := range req.Data {
		for k := range row {
			colsMap[k] = true
		}
	}
	colsMap["id_instancia"] = true

	// Converter para slice e ordenar
	var cols []string
	for col := range colsMap {
		if !query.IsValidColumnName(col) {
			return 0, fmt.Errorf("%w: %s", models.ErrInvalidColumn, col)
		}
		cols = append(cols, col)
	}
	sort.Strings(cols)

	// ✅ ÚNICA MUDANÇA: MySQL usa ? ao invés de $1, $2, $3
	var valuePlaceholders []string
	var allValues []interface{}

	for _, row := range req.Data {
		row["id_instancia"] = req.InstanceID

		// Gerar placeholders MySQL: (?, ?, ?)
		var rowPlaceholders []string
		for range cols {
			rowPlaceholders = append(rowPlaceholders, "?") // ✅ MySQL
		}
		valuePlaceholders = append(valuePlaceholders, "("+strings.Join(rowPlaceholders, ",")+")")

		// Adicionar valores na ordem das colunas
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

	fmt.Printf("📝 BATCH SQL: %s\n", queryStr)
	fmt.Printf("📊 BATCH Args: %v\n", allValues)

	_, err = config.MasterDB.Exec(queryStr, allValues...)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", models.ErrInsertFailed, err)
	}

	return len(req.Data), nil
}
