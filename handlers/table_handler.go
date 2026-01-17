package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/models"
	
	"fmt"
	"strings"

	tableService "meu-provedor/services/table"
)

func CreateProjectTable(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	projectCode := r.URL.Query().Get("project_code")
	tableName, err := tableService.Create(projectCode, req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "TABLE CREATED",
		"table":   tableName,
	})
}

func ListProjectTables(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tables, err := tableService.List(projectCode)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(tables)
}

func DeleteProjectTable(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tableName := r.URL.Query().Get("table")

	if err := tableService.Delete(projectCode, tableName); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("TABLE DELETED"))
}




// GET TABLE DETAILS
func GetTableDetails(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tableName := r.URL.Query().Get("table")
	if projectCode == "" || tableName == "" {
		http.Error(w, "project_code and table are required", 400)
		return
	}

	details, err := tableService.GetDetails(projectCode, tableName)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(details)
}

// ADD COLUMN
func AddColumn(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tableName := r.URL.Query().Get("table")

	var col tableService.ColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	if err := tableService.AddColumn(projectCode, tableName, col); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("COLUMN ADDED"))
}

// MODIFY COLUMN
func ModifyColumn(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tableName := r.URL.Query().Get("table")

	var col tableService.ColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	if err := tableService.ModifyColumn(projectCode, tableName, col); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("COLUMN MODIFIED"))
}

// DROP COLUMN
func DropColumn(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tableName := r.URL.Query().Get("table")
	columnName := r.URL.Query().Get("column")

	if err := tableService.DropColumn(projectCode, tableName, columnName); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("COLUMN DROPPED"))
}

// ADD INDEX
func AddIndex(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tableName := r.URL.Query().Get("table")

	var idx tableService.IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&idx); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	if err := tableService.AddIndex(projectCode, tableName, idx); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("INDEX ADDED"))
}

// DROP INDEX
func DropIndex(w http.ResponseWriter, r *http.Request) {
	projectCode := r.URL.Query().Get("project_code")
	tableName := r.URL.Query().Get("table")
	indexName := r.URL.Query().Get("index")

	if err := tableService.DropIndex(projectCode, tableName, indexName); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("INDEX DROPPED"))
}

