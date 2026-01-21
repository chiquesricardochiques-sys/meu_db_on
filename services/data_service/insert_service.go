package services

import (
	"fmt"
	"log"
	"meu-provedor/config"
	"meu-provedor/engine/query"
	"meu-provedor/models"
)

// ExecuteInsert executa INSERT único com validação completa
func ExecuteInsert(req models.InsertRequest) (int64, error) {
	// ✅ PASSO 1: Validar requisição
	if err := req.Validate(); err != nil {
		return 0, fmt.Errorf("validação falhou: %w", err)
	}
	
	// ✅ PASSO 2: Buscar código do projeto
	projectCode, err := config.GetProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, fmt.Errorf("projeto não encontrado: %w", err)
	}
	
	// ✅ PASSO 3: Construir nome da tabela
	tableName := fmt.Sprintf("%s_%s", projectCode, req.Table)
	
	// ✅ PASSO 4: Preparar colunas e valores
	columns := make([]string, 0, len(req.Columns)+1)
	values := make([]interface{}, 0, len(req.Columns)+1)
	
	// Adicionar id_instancia primeiro
	columns = append(columns, "id_instancia")
	values = append(values, req.InstanceID)
	
	// Adicionar colunas do request
	for _, col := range req.Columns {
		columns = append(columns, col.Name)
		values = append(values, col.Value)
	}
	
	// ✅ PASSO 5: Construir query
	builder := query.NewInsert(tableName).SetColumns(columns)
	if err := builder.AddRow(values); err != nil {
		return 0, fmt.Errorf("erro ao adicionar row: %w", err)
	}
	
	sqlQuery, args, err := builder.Build()
	if err != nil {
		return 0, fmt.Errorf("erro ao construir SQL: %w", err)
	}
	
	// ✅ PASSO 6: Log para debug
	log.Printf("📝 SQL: %s", sqlQuery)
	log.Printf("📊 Args: %v", args)
	
	// ✅ PASSO 7: Executar
	result, err := config.MasterDB.Exec(sqlQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("erro ao executar INSERT: %w", err)
	}
	
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("erro ao obter ID: %w", err)
	}
	
	log.Printf("✅ Registro inserido com ID: %d", lastID)
	return lastID, nil
}

// ExecuteBatchInsert executa múltiplos INSERTs em uma única query
func ExecuteBatchInsert(req models.BatchInsertRequest) (int, error) {
	// ✅ PASSO 1: Validar requisição
	if err := req.Validate(); err != nil {
		return 0, fmt.Errorf("validação falhou: %w", err)
	}
	
	// ✅ PASSO 2: Buscar código do projeto
	projectCode, err := config.GetProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, fmt.Errorf("projeto não encontrado: %w", err)
	}
	
	// ✅ PASSO 3: Construir nome da tabela
	tableName := fmt.Sprintf("%s_%s", projectCode, req.Table)
	
	// ✅ PASSO 4: Extrair nomes das colunas da primeira row
	// Assumindo que todas as rows têm as mesmas colunas
	firstRow := req.Rows[0]
	columns := make([]string, 0, len(firstRow)+1)
	
	// Adicionar id_instancia primeiro
	columns = append(columns, "id_instancia")
	
	// Adicionar colunas da primeira row
	for _, col := range firstRow {
		columns = append(columns, col.Name)
	}
	
	// ✅ PASSO 5: Construir query
	builder := query.NewInsert(tableName).SetColumns(columns)
	
	// Adicionar cada row
	for _, row := range req.Rows {
		values := make([]interface{}, 0, len(row)+1)
		
		// Adicionar id_instancia
		values = append(values, req.InstanceID)
		
		// Adicionar valores da row
		for _, col := range row {
			values = append(values, col.Value)
		}
		
		if err := builder.AddRow(values); err != nil {
			return 0, fmt.Errorf("erro ao adicionar row: %w", err)
		}
	}
	
	sqlQuery, args, err := builder.Build()
	if err != nil {
		return 0, fmt.Errorf("erro ao construir SQL: %w", err)
	}
	
	// ✅ PASSO 6: Log para debug
	log.Printf("📝 BATCH SQL: %s", sqlQuery)
	log.Printf("📊 BATCH Args: %v", args)
	
	// ✅ PASSO 7: Executar
	_, err = config.MasterDB.Exec(sqlQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("erro ao executar BATCH INSERT: %w", err)
	}
	
	log.Printf("✅ %d registros inseridos", len(req.Rows))
	return len(req.Rows), nil
}
