package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"meu-provedor/config"
)

/*
========================
STRUCTS
========================
*/

// Projeto
type ProjectRequest struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	ApiKey  string `json:"api_key"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// Instância
type InstanceRequest struct {
	ProjectID   int64                  `json:"project_id"`
	Name        string                 `json:"name"`
	Code        string                 `json:"code"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Settings    map[string]interface{} `json:"settings"`
}

/*
========================
PROJETOS - CRUD
========================
*/

// CREATE PROJECT
func CreateProject(w http.ResponseWriter, r *http.Request) {
	var req ProjectRequest
	json.NewDecoder(r.Body).Decode(&req)

	query := `
		INSERT INTO projects (name, code, api_key, type, version, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := config.MasterDB.Exec(
		query,
		req.Name,
		req.Code,
		req.ApiKey,
		req.Type,
		req.Version,
		req.Status,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("PROJECT CREATED"))
}

// LIST PROJECTS
func ListProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := config.MasterDB.Query(`
		SELECT id, name, code, type, version, status
		FROM projects
	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var (
			id      int64
			name    string
			code    string
			ptype   string
			version string
			status  string
		)

		rows.Scan(&id, &name, &code, &ptype, &version, &status)

		result = append(result, map[string]interface{}{
			"id":      id,
			"name":    name,
			"code":    code,
			"type":    ptype,
			"version": version,
			"status":  status,
		})
	}

	json.NewEncoder(w).Encode(result)
}

// UPDATE PROJECT
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var req ProjectRequest
	json.NewDecoder(r.Body).Decode(&req)

	query := `
		UPDATE projects
		SET name = ?, code = ?, type = ?, version = ?, status = ?
		WHERE id = ?
	`

	_, err := config.MasterDB.Exec(
		query,
		req.Name,
		req.Code,
		req.Type,
		req.Version,
		req.Status,
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("PROJECT UPDATED"))
}

// DELETE PROJECT
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := config.MasterDB.Exec(
		"DELETE FROM projects WHERE id = ?",
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("PROJECT DELETED"))
}

/*
========================
INSTÂNCIAS - CRUD
========================
*/

// CREATE INSTANCE
func CreateInstance(w http.ResponseWriter, r *http.Request) {
	var req InstanceRequest
	json.NewDecoder(r.Body).Decode(&req)

	settingsJSON, _ := json.Marshal(req.Settings)

	query := `
		INSERT INTO instancias_projetion
		(project_id, name, code, description, status, settings)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := config.MasterDB.Exec(
		query,
		req.ProjectID,
		req.Name,
		req.Code,
		req.Description,
		req.Status,
		settingsJSON,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("INSTANCE CREATED"))
}

// LIST INSTANCES BY PROJECT
func ListInstances(w http.ResponseWriter, r *http.Request) {
	projectID, _ := strconv.Atoi(mux.Vars(r)["project_id"])

	rows, err := config.MasterDB.Query(`
		SELECT id, name, code, description, status, settings
		FROM instancias_projetion
		WHERE project_id = ?
	`, projectID)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var result []map[string]interface{}

	for rows.Next() {
		var (
			id          int64
			name        string
			code        string
			description sql.NullString
			status      string
			settings    []byte
		)

		rows.Scan(&id, &name, &code, &description, &status, &settings)

		var settingsMap map[string]interface{}
		json.Unmarshal(settings, &settingsMap)

		result = append(result, map[string]interface{}{
			"id":          id,
			"name":        name,
			"code":        code,
			"description": description.String,
			"status":      status,
			"settings":    settingsMap,
		})
	}

	json.NewEncoder(w).Encode(result)
}

// UPDATE INSTANCE
func UpdateInstance(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var req InstanceRequest
	json.NewDecoder(r.Body).Decode(&req)

	settingsJSON, _ := json.Marshal(req.Settings)

	query := `
		UPDATE instancias_projetion
		SET name = ?, code = ?, description = ?, status = ?, settings = ?
		WHERE id = ?
	`

	_, err := config.MasterDB.Exec(
		query,
		req.Name,
		req.Code,
		req.Description,
		req.Status,
		settingsJSON,
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("INSTANCE UPDATED"))
}

// DELETE INSTANCE
func DeleteInstance(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := config.MasterDB.Exec(
		"DELETE FROM instancias_projetion WHERE id = ?",
		id,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("INSTANCE DELETED"))
}
