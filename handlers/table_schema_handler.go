package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"meu-provedor/config"
)

/*
====================================================
VALIDAÇÕES
====================================================
*/

var validName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateName(name string) bool {
	return validName.MatchString(name)
}

/*
====================================================
STRUCTS
====================================================
*/

type ColumnRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Unique   bool   `json:"unique"`
}

type IndexRequest struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"` // UNIQUE ou INDEX
}

type CreateTableRequest struct {
	ProjectID int64           `json:"project_id"`
	TableName string          `json:"table_name"`
	Columns   []ColumnRequest `json:"columns"`
	Indexes   []IndexRequest  `json:"indexes,omitempty"`
}

type ColumnDetail struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Nullable bool        `json:"nullable"`
	Default  interface{} `json:"default"`
	Key      string      `json:"key"`
	Extra    string      `json:"extra"`
}

type IndexDetail struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"`
}

type TableDetail struct {
	Name    string         `json:"name"`
	Columns []ColumnDetail `json:"columns"`
	Indexes []IndexDetail  `json:"indexes"`
}

/*
====================================================
HELPERS
====================================================
*/

func getProjectCode(projectID int64) (string, error) {
	var code string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&code)
	return code, err
}

func getTableColumns(tableName string) ([]ColumnDetail, error) {
	rows, err := config.MasterDB.Query(fmt.Sprintf("DESCRIBE %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnDetail

	for rows.Next() {
		var field, colType, null, key, extra string
		var defaultVal sql.NullString

		err := rows.Scan(&field, &colType, &null, &key, &defaultVal, &extra)
		if err != nil {
			continue
		}

		var defValue interface{}
		if defaultVal.Valid {
			defValue = defaultVal.String
		}

		columns = append(columns, ColumnDetail{
			Name:     field,
			Type:     colType,
			Nullable: null == "YES",
			Default:  defValue,
			Key:      key,
			Extra:    extra,
		})
	}

	return columns, nil
}

func getTableIndexes(tableName string) ([]IndexDetail, error) {
	rows, err := config.MasterDB.Query(fmt.Sprintf("SHOW INDEX FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexMap := make(map[string]*IndexDetail)

	for rows.Next() {
		var table, keyName, columnName, collation, indexType, comment, indexComment string
		var nonUnique, seqInIndex, cardinality int
		var subPart, packed, null, visible sql.NullString

		err := rows.Scan(
			&table, &nonUnique, &keyName, &seqInIndex, &columnName,
			&collation, &cardinality, &subPart, &packed, &null,
			&indexType, &comment, &indexComment, &visible,
		)
		if err != nil {
			continue
		}

		if _, exists := indexMap[keyName]; !exists {
			idxType := "INDEX"
			if keyName == "PRIMARY" {
				idxType = "PRIMARY"
			} else if nonUnique == 0 {
				idxType = "UNIQUE"
			}

			indexMap[keyName] = &IndexDetail{
				Name:    keyName,
				Columns: []string{},
				Type:    idxType,
			}
		}

		indexMap[keyName].Columns = append(indexMap[keyName].Columns, columnName)
	}

	var indexes []IndexDetail
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}

	return indexes, nil
}

/*
====================================================
CREATE TABLE (UNIFICADA - ACEITA COM OU SEM ÍNDICES)
====================================================
*/

func CreateProjectTable(w http.ResponseWriter, r *http.Request) {
	var req CreateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if !validateName(req.TableName) {
		http.Error(w, "Invalid table name", 400)
		return
	}

	projectCode, err := getProjectCode(req.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTableName := fmt.Sprintf("%s_%s", projectCode, req.TableName)

	var columns []string

	// Colunas padrão obrigatórias
	columns = append(columns, "id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY")
	columns = append(columns, `id_instancia BIGINT UNSIGNED NOT NULL,
		FOREIGN KEY (id_instancia)
		REFERENCES instancias_projetion(id)
		ON DELETE CASCADE`)

	// Colunas personalizadas
	for _, col := range req.Columns {
		if !validateName(col.Name) {
			http.Error(w, "Invalid column name: "+col.Name, 400)
			return
		}

		def := col.Name + " " + col.Type
		if !col.Nullable {
			def += " NOT NULL"
		}
		if col.Unique {
			def += " UNIQUE"
		}

		columns = append(columns, def)
	}

	// Índices opcionais
	for _, idx := range req.Indexes {
		if len(idx.Columns) == 0 {
			continue
		}

		if idx.Name != "" && !validateName(idx.Name) {
			http.Error(w, "Invalid index name: "+idx.Name, 400)
			return
		}

		for _, col := range idx.Columns {
			if !validateName(col) {
				http.Error(w, "Invalid column in index: "+col, 400)
				return
			}
		}

		idxDef := ""
		if idx.Type == "UNIQUE" {
			idxDef = fmt.Sprintf("UNIQUE KEY %s (%s)", idx.Name, strings.Join(idx.Columns, ","))
		} else {
			idxDef = fmt.Sprintf("INDEX %s (%s)", idx.Name, strings.Join(idx.Columns, ","))
		}
		columns = append(columns, idxDef)
	}

	createSQL := fmt.Sprintf("CREATE TABLE %s (%s)", fullTableName, strings.Join(columns, ","))

	if _, err := config.MasterDB.Exec(createSQL); err != nil {
		http.Error(w, "Error creating table: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "TABLE CREATED",
		"table":   fullTableName,
	})
}

/*
====================================================
LIST TABLES (COM DETALHES OPCIONAIS)
====================================================
*/

func ListProjectTables(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	detailed := r.URL.Query().Get("detailed") // ?detailed=true para detalhes

	if projectID == "" {
		http.Error(w, "project_id required", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	rows, err := config.MasterDB.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		AND table_name LIKE ?
	`, projectCode+"_%")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	// Se não quer detalhes, retorna só os nomes
	if detailed != "true" {
		var tables []string
		for rows.Next() {
			var fullName string
			rows.Scan(&fullName)
			displayName := strings.TrimPrefix(fullName, projectCode+"_")
			tables = append(tables, displayName)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tables)
		return
	}

	// Retorna com detalhes completos
	var tables []TableDetail

	for rows.Next() {
		var fullTableName string
		rows.Scan(&fullTableName)

		columns, err := getTableColumns(fullTableName)
		if err != nil {
			continue
		}

		indexes, err := getTableIndexes(fullTableName)
		if err != nil {
			indexes = []IndexDetail{}
		}

		displayName := strings.TrimPrefix(fullTableName, projectCode+"_")

		tables = append(tables, TableDetail{
			Name:    displayName,
			Columns: columns,
			Indexes: indexes,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tables)
}

/*
====================================================
GET TABLE DETAILS
====================================================
*/

func GetTableDetails(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	table := r.URL.Query().Get("table")

	if projectID == "" || table == "" {
		http.Error(w, "project_id and table required", 400)
		return
	}

	if !validateName(table) {
		http.Error(w, "Invalid table name", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTableName := fmt.Sprintf("%s_%s", projectCode, table)

	columns, err := getTableColumns(fullTableName)
	if err != nil {
		http.Error(w, "Error getting columns: "+err.Error(), 500)
		return
	}

	indexes, err := getTableIndexes(fullTableName)
	if err != nil {
		indexes = []IndexDetail{}
	}

	tableDetail := TableDetail{
		Name:    table,
		Columns: columns,
		Indexes: indexes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tableDetail)
}

/*
====================================================
DELETE TABLE
====================================================
*/

func DeleteProjectTable(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	table := r.URL.Query().Get("table")

	if projectID == "" || table == "" {
		http.Error(w, "project_id and table required", 400)
		return
	}

	if !validateName(table) {
		http.Error(w, "Invalid table name", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTable := fmt.Sprintf("%s_%s", projectCode, table)

	if _, err := config.MasterDB.Exec("DROP TABLE " + fullTable); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("TABLE DELETED"))
}

/*
====================================================
ADD COLUMN
====================================================
*/

func AddColumn(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	table := r.URL.Query().Get("table")

	var col ColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if !validateName(col.Name) || !validateName(table) {
		http.Error(w, "Invalid name", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTable := fmt.Sprintf("%s_%s", projectCode, table)

	def := col.Name + " " + col.Type
	if !col.Nullable {
		def += " NOT NULL"
	}
	if col.Unique {
		def += " UNIQUE"
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", fullTable, def)

	if _, err := config.MasterDB.Exec(query); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("COLUMN ADDED"))
}

/*
====================================================
DROP COLUMN
====================================================
*/

func DropColumn(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	table := r.URL.Query().Get("table")
	column := r.URL.Query().Get("column")

	if !validateName(table) || !validateName(column) {
		http.Error(w, "Invalid name", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTable := fmt.Sprintf("%s_%s", projectCode, table)

	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", fullTable, column)

	if _, err := config.MasterDB.Exec(query); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("COLUMN DROPPED"))
}

/*
====================================================
MODIFY COLUMN
====================================================
*/

func ModifyColumn(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	table := r.URL.Query().Get("table")

	var col ColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if !validateName(col.Name) || !validateName(table) {
		http.Error(w, "Invalid name", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTable := fmt.Sprintf("%s_%s", projectCode, table)

	def := col.Name + " " + col.Type
	if !col.Nullable {
		def += " NOT NULL"
	}
	if col.Unique {
		def += " UNIQUE"
	}

	query := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", fullTable, def)

	if _, err := config.MasterDB.Exec(query); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("COLUMN MODIFIED"))
}

/*
====================================================
ADD INDEX
====================================================
*/

func AddIndex(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	table := r.URL.Query().Get("table")

	var idx IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&idx); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if !validateName(table) {
		http.Error(w, "Invalid table name", 400)
		return
	}

	if idx.Name != "" && !validateName(idx.Name) {
		http.Error(w, "Invalid index name", 400)
		return
	}

	if len(idx.Columns) == 0 {
		http.Error(w, "At least one column required", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTable := fmt.Sprintf("%s_%s", projectCode, table)

	var query string
	if idx.Type == "UNIQUE" {
		query = fmt.Sprintf("ALTER TABLE %s ADD UNIQUE INDEX %s (%s)",
			fullTable, idx.Name, strings.Join(idx.Columns, ","))
	} else {
		query = fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s)",
			fullTable, idx.Name, strings.Join(idx.Columns, ","))
	}

	if _, err := config.MasterDB.Exec(query); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("INDEX ADDED"))
}

/*
====================================================
DROP INDEX
====================================================
*/

func DropIndex(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	table := r.URL.Query().Get("table")
	indexName := r.URL.Query().Get("index")

	if !validateName(table) || !validateName(indexName) {
		http.Error(w, "Invalid name", 400)
		return
	}

	var projectCode string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	fullTable := fmt.Sprintf("%s_%s", projectCode, table)

	query := fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", fullTable, indexName)

	if _, err := config.MasterDB.Exec(query); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("INDEX DROPPED"))
}
