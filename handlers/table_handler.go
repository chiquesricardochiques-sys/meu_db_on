package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/models"
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
