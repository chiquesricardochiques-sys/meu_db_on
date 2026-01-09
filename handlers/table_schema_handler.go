package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"meu-provedor/config"
)

/*
====================================================
VALIDAÇÕES BÁSICAS
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

type CreateTableRequest struct {
	ProjectID int64           `json:"project_id"`
	TableName string          `json:"table_name"`
	Columns   []ColumnRequest `json:"columns"`
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

/*
====================================================
CREATE TABLE
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

	// ID padrão
	columns = append(columns, "id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY")

	// Separação por instância (OBRIGATÓRIO)
	columns = append(columns, `
		id_instancia BIGINT UNSIGNED NOT NULL,
		FOREIGN KEY (id_instancia)
		REFERENCES instancias_projetion(id)
		ON DELETE CASCADE
	`)

	for _, col := range req.Columns {
		if !validateName(col.Name) {
			http.Error(w, "Invalid column name", 400)
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

	createSQL := fmt.Sprintf(
		"CREATE TABLE %s (%s)",
		fullTableName,
		strings.Join(columns, ","),
	)

	if _, err := config.MasterDB.Exec(createSQL); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("TABLE CREATED"))
}

/*
====================================================
LIST TABLES BY PROJECT
====================================================
*/

func ListProjectTables(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
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

	var tables []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}

	json.NewEncoder(w).Encode(tables)
}

/*
====================================================
DROP TABLE
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
	config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

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


// GetQueryString retorna o valor de uma query string ou valor default
func GetQueryString(w http.ResponseWriter, r *http.Request) {
    apiKey := r.URL.Query().Get("api_key")
    project, err := security.ValidateApiKey(apiKey)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    db, err := config.GetDBConnection(project)
    if err != nil {
        http.Error(w, "Erro ao conectar banco do projeto", http.StatusInternalServerError)
        return
    }

    table := r.URL.Query().Get("table")
    if table == "" {
        http.Error(w, "É necessário fornecer o parâmetro 'table'", http.StatusBadRequest)
        return
    }

    rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
    if err != nil {
        http.Error(w, "Erro ao consultar dados", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    cols, _ := rows.Columns()
    results := []map[string]interface{}{}

    for rows.Next() {
        columns := make([]interface{}, len(cols))
        columnPointers := make([]interface{}, len(cols))
        for i := range columns {
            columnPointers[i] = &columns[i]
        }

        if err := rows.Scan(columnPointers...); err != nil {
            continue
        }

        m := make(map[string]interface{})
        for i, colName := range cols {
            val := columnPointers[i].(*interface{})
            m[colName] = *val
        }
        results = append(results, m)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
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
	config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ?",
		projectID,
	).Scan(&projectCode)

	fullTable := fmt.Sprintf("%s_%s", projectCode, table)

	query := fmt.Sprintf(
		"ALTER TABLE %s DROP COLUMN %s",
		fullTable,
		column,
	)

	if _, err := config.MasterDB.Exec(query); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("COLUMN DROPPED"))
}



