package data_service

import (
	"fmt"
	"strings"

	"meu-provedor/config"
	"meu-provedor/handlers"
)

func ExecuteBatchInsert(req handlers.BatchInsertRequest) (int, error) {

	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, fmt.Errorf("project not found")
	}

	table := fmt.Sprintf("%s_%s", projectCode, req.Table)

	// Coletar todas as colunas únicas
	colsMap := make(map[string]bool)
	for _, row := range req.Data {
		for k := range row {
			colsMap[k] = true
		}
	}
	colsMap["id_instancia"] = true // obrigatória
	var cols []string
	for col := range colsMap {
		cols = append(cols, col)
	}

	placeholders := "(" + strings.Repeat("?,", len(cols)-1) + "?)"
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

	_, err = config.MasterDB.Exec(queryStr, allValues...)
	if err != nil {
		return 0, fmt.Errorf("Batch insert failed: %v", err)
	}

	return len(req.Data), nil
}
