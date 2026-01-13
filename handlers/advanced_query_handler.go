package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"meu-provedor/config"
)

/*
====================================================
STRUCTS PARA QUERIES AVANÇADAS
====================================================
*/

type JoinConfig struct {
	Type       string `json:"type"`        // INNER, LEFT, RIGHT
	Table      string `json:"table"`       // Tabela a juntar
	On         string `json:"on"`          // Condição: "tabela1.id = tabela2.fk"
	Alias      string `json:"alias"`       // Apelido da tabela
}

type AdvancedQueryRequest struct {
	ProjectID   int64                    `json:"project_id"`
	InstanceID  int64                    `json:"id_instancia"`
	Table       string                   `json:"table"`
	Alias       string                   `json:"alias,omitempty"`       // Apelido da tabela principal
	Select      []string                 `json:"select,omitempty"`      // Colunas a selecionar
	Joins       []JoinConfig             `json:"joins,omitempty"`       // Configuração de JOINs
	Where       map[string]interface{}   `json:"where,omitempty"`       // Filtros simples (AND)
	WhereRaw    string                   `json:"where_raw,omitempty"`   // WHERE customizado
	OrderBy     string                   `json:"order_by,omitempty"`    // Ordenação
	GroupBy     string                   `json:"group_by,omitempty"`    // Agrupamento
	Having      string                   `json:"having,omitempty"`      // HAVING para GROUP BY
	Limit       int                      `json:"limit,omitempty"`       // Limite de resultados
	Offset      int                      `json:"offset,omitempty"`      // Offset para paginação
}

type BatchInsertRequest struct {
	ProjectID  int64                    `json:"project_id"`
	InstanceID int64                    `json:"id_instancia"`
	Table      string                   `json:"table"`
	Data       []map[string]interface{} `json:"data"` // Array de registros
}

type BatchUpdateRequest struct {
	ProjectID  int64                    `json:"project_id"`
	InstanceID int64                    `json:"id_instancia"`
	Table      string                   `json:"table"`
	Updates    []struct {
		Data  map[string]interface{} `json:"data"`
		Where map[string]interface{} `json:"where"`
	} `json:"updates"`
}

/*
====================================================
ADVANCED SELECT - SELECT COM JOINS E FILTROS AVANÇADOS
====================================================
*/

func AdvancedSelect(w http.ResponseWriter, r *http.Request) {
	var req AdvancedQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	mainTable, err := buildTableName(projectCode, req.Table)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Alias da tabela principal
	mainAlias := req.Alias
	if mainAlias == "" {
		mainAlias = mainTable
	}

	// SELECT
	selectClause := "*"
	if len(req.Select) > 0 {
		selectClause = strings.Join(req.Select, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s AS %s", selectClause, mainTable, mainAlias)

	// JOINS
	for _, join := range req.Joins {
		if !isValidIdentifier(join.Table) {
			http.Error(w, "Invalid join table name", 400)
			return
		}

		joinTable, err := buildTableName(projectCode, join.Table)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		joinAlias := join.Alias
		if joinAlias == "" {
			joinAlias = joinTable
		}

		joinType := strings.ToUpper(join.Type)
		if joinType == "" {
			joinType = "INNER"
		}

		query += fmt.Sprintf(" %s JOIN %s AS %s ON %s", joinType, joinTable, joinAlias, join.On)
	}

	// WHERE
	var whereConditions []string
	var values []interface{}

	// Filtro obrigatório: id_instancia
	whereConditions = append(whereConditions, fmt.Sprintf("%s.id_instancia = ?", mainAlias))
	values = append(values, req.InstanceID)

	// Filtros simples (AND)
	for k, v := range req.Where {
		if !isValidIdentifier(k) {
			http.Error(w, "Invalid filter column", 400)
			return
		}
		whereConditions = append(whereConditions, k+" = ?")
		values = append(values, v)
	}

	// WHERE customizado (cuidado: pode causar SQL injection se não validado)
	if req.WhereRaw != "" {
		whereConditions = append(whereConditions, "("+req.WhereRaw+")")
	}

	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// GROUP BY
	if req.GroupBy != "" {
		query += " GROUP BY " + req.GroupBy
	}

	// HAVING
	if req.Having != "" {
		query += " HAVING " + req.Having
	}

	// ORDER BY
	if req.OrderBy != "" {
		query += " ORDER BY " + req.OrderBy
	}

	// LIMIT e OFFSET
	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", req.Limit)
		if req.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", req.Offset)
		}
	}

	// Executar query
	rows, err := config.MasterDB.Query(query, values...)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), 500)
		return
	}
	defer rows.Close()

	// Converter para JSON
	result := rowsToMap(rows)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

/*
====================================================
BATCH INSERT - INSERIR MÚLTIPLOS REGISTROS
====================================================
*/

func BatchInsert(w http.ResponseWriter, r *http.Request) {
	var req BatchInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if len(req.Data) == 0 {
		http.Error(w, "No data provided", 400)
		return
	}

	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	table, err := buildTableName(projectCode, req.Table)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Coletar todas as colunas únicas
	colsMap := make(map[string]bool)
	for _, row := range req.Data {
		for k := range row {
			if !isValidIdentifier(k) {
				http.Error(w, "Invalid column name: "+k, 400)
				return
			}
			colsMap[k] = true
		}
	}

	colsMap["id_instancia"] = true
	var cols []string
	for col := range colsMap {
		cols = append(cols, col)
	}

	// Preparar query
	placeholders := "(" + strings.Repeat("?,", len(cols)-1) + "?)"
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", table, strings.Join(cols, ","))

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

	query += strings.Join(valuePlaceholders, ",")

	_, err = config.MasterDB.Exec(query, allValues...)
	if err != nil {
		http.Error(w, "Batch insert failed: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Batch insert completed",
		"count":   len(req.Data),
	})
}

/*
====================================================
BATCH UPDATE - ATUALIZAR MÚLTIPLOS REGISTROS
====================================================
*/

func BatchUpdate(w http.ResponseWriter, r *http.Request) {
	var req BatchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if len(req.Updates) == 0 {
		http.Error(w, "No updates provided", 400)
		return
	}

	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	table, err := buildTableName(projectCode, req.Table)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	tx, err := config.MasterDB.Begin()
	if err != nil {
		http.Error(w, "Transaction error", 500)
		return
	}

	count := 0
	for _, upd := range req.Updates {
		var sets []string
		var values []interface{}

		for k, v := range upd.Data {
			if !isValidIdentifier(k) {
				tx.Rollback()
				http.Error(w, "Invalid column: "+k, 400)
				return
			}
			sets = append(sets, k+" = ?")
			values = append(values, v)
		}

		var whereConditions []string
		whereConditions = append(whereConditions, "id_instancia = ?")
		values = append(values, req.InstanceID)

		for k, v := range upd.Where {
			if !isValidIdentifier(k) {
				tx.Rollback()
				http.Error(w, "Invalid where column: "+k, 400)
				return
			}
			whereConditions = append(whereConditions, k+" = ?")
			values = append(values, v)
		}

		query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
			table,
			strings.Join(sets, ", "),
			strings.Join(whereConditions, " AND "))

		_, err := tx.Exec(query, values...)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Update failed: "+err.Error(), 500)
			return
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Commit failed", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Batch update completed",
		"count":   count,
	})
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