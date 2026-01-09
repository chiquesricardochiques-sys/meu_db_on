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
VALIDAÇÕES
====================================================
*/

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isValidIdentifier(v string) bool {
	return validIdentifier.MatchString(v)
}

/*
====================================================
STRUCTS
====================================================
*/

type GenericRequest struct {
	ProjectID   int64                  `json:"project_id"`
	InstanceID  int64                  `json:"id_instancia"`
	Table       string                 `json:"table"`
	Data        map[string]interface{} `json:"data"`
	Filters     map[string]interface{} `json:"filters"`
	ID          int64                  `json:"id"`
}

/*
====================================================
HELPERS
====================================================
*/

func getProjectCodeByID(projectID int64) (string, error) {
	var code string
	err := config.MasterDB.QueryRow(
		"SELECT code FROM projects WHERE id = ? AND status = 'active'",
		projectID,
	).Scan(&code)

	return code, err
}

func buildTableName(projectCode, table string) (string, error) {
	if !isValidIdentifier(table) {
		return "", fmt.Errorf("invalid table name")
	}
	return projectCode + "_" + table, nil
}

/*
====================================================
INSERT
====================================================
*/

func Insert(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
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

	req.Data["id_instancia"] = req.InstanceID

	var cols []string
	var placeholders []string
	var values []interface{}

	for k, v := range req.Data {
		if !isValidIdentifier(k) {
			http.Error(w, "Invalid column name", 400)
			return
		}
		cols = append(cols, k)
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ","),
		strings.Join(placeholders, ","),
	)

	if _, err := config.MasterDB.Exec(query, values...); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("INSERT OK"))
}

/*
====================================================
GET (LIST)
====================================================
*/

func Get(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		http.Error(w, "Project not found", 404)
		return
	}

	table, _ := buildTableName(projectCode, req.Table)

	var where []string
	var values []interface{}

	where = append(where, "id_instancia = ?")
	values = append(values, req.InstanceID)

	for k, v := range req.Filters {
		if !isValidIdentifier(k) {
			http.Error(w, "Invalid filter", 400)
			return
		}
		where = append(where, k+" = ?")
		values = append(values, v)
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s",
		table,
		strings.Join(where, " AND "),
	)

	rows, err := config.MasterDB.Query(query, values...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

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
			row[col] = values[i]
		}
		result = append(result, row)
	}

	json.NewEncoder(w).Encode(result)
}

/*
====================================================
UPDATE
====================================================
*/

func Update(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	json.NewDecoder(r.Body).Decode(&req)

	projectCode, _ := getProjectCodeByID(req.ProjectID)
	table, _ := buildTableName(projectCode, req.Table)

	var sets []string
	var values []interface{}

	for k, v := range req.Data {
		if !isValidIdentifier(k) {
			http.Error(w, "Invalid column", 400)
			return
		}
		sets = append(sets, k+" = ?")
		values = append(values, v)
	}

	values = append(values, req.ID, req.InstanceID)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ? AND id_instancia = ?",
		table,
		strings.Join(sets, ","),
	)

	if _, err := config.MasterDB.Exec(query, values...); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("UPDATE OK"))
}

/*
====================================================
DELETE
====================================================
*/

func Delete(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	json.NewDecoder(r.Body).Decode(&req)

	projectCode, _ := getProjectCodeByID(req.ProjectID)
	table, _ := buildTableName(projectCode, req.Table)

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE id = ? AND id_instancia = ?",
		table,
	)

	if _, err := config.MasterDB.Exec(query, req.ID, req.InstanceID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("DELETE OK"))
}

