package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"meu-provedor/config"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Estrutura para criar/atualizar projeto
type ProjectRequest struct {
	Name     string `json:"name"`
	Database string `json:"database"`
	ApiKey   string `json:"api_key"`
}

// Estrutura para criar tabela
type TableRequest struct {
	TableName string            `json:"table_name"`
	Columns   map[string]string `json:"columns"` // nome da coluna -> tipo
}

// ===================== PROJETOS =====================

// Cria novo projeto
func CreateProject(w http.ResponseWriter, r *http.Request) {
	var req ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	// Cria projeto no banco master
	sqlQuery := "INSERT INTO projects (name, database_name, api_key) VALUES (?, ?, ?)"
	_, err := config.MasterDB.Exec(sqlQuery, req.Name, req.Database, req.ApiKey)
	if err != nil {
		http.Error(w, "Erro ao criar projeto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Cria banco físico do projeto
	_, err = config.MasterDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", req.Database))
	if err != nil {
		http.Error(w, "Erro ao criar banco do projeto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Projeto criado com sucesso!"))
}

// Lista todos os projetos
func ListProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := config.MasterDB.Query("SELECT id, name, database_name, api_key FROM projects")
	if err != nil {
		http.Error(w, "Erro ao buscar projetos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	projects := []config.Project{}
	for rows.Next() {
		var p config.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Database, &p.ApiKey); err != nil {
			continue
		}
		projects = append(projects, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// Atualiza projeto
func UpdateProject(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var req ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	sqlQuery := "UPDATE projects SET name = ?, database_name = ?, api_key = ? WHERE id = ?"
	_, err := config.MasterDB.Exec(sqlQuery, req.Name, req.Database, req.ApiKey, id)
	if err != nil {
		http.Error(w, "Erro ao atualizar projeto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Projeto atualizado!"))
}

// Deleta projeto
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var dbName string
	row := config.MasterDB.QueryRow("SELECT database_name FROM projects WHERE id = ?", id)
	row.Scan(&dbName)

	// Deleta projeto do master
	_, err := config.MasterDB.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Erro ao deletar projeto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Deleta banco físico
	_, err = config.MasterDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		log.Println("⚠️ Erro ao deletar banco físico:", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Projeto deletado!"))
}

// ===================== TABELAS =====================

// Cria nova tabela no projeto
func CreateTable(w http.ResponseWriter, r *http.Request) {
	projectIDStr := mux.Vars(r)["id"]
	projectID, _ := strconv.Atoi(projectIDStr)

	// Busca projeto
	var project config.Project
	row := config.MasterDB.QueryRow("SELECT id, name, database_name, api_key FROM projects WHERE id = ?", projectID)
	row.Scan(&project.ID, &project.Name, &project.Database, &project.ApiKey)

	var req TableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	// Monta query de criação de tabela
	columns := ""
	i := 0
	for col, tipo := range req.Columns {
		if i > 0 {
			columns += ", "
		}
		columns += fmt.Sprintf("%s %s", col, tipo)
		i++
	}
	sqlQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", req.TableName, columns)

	db, err := config.GetDBConnection(&project)
	if err != nil {
		http.Error(w, "Erro ao conectar banco do projeto", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(sqlQuery)
	if err != nil {
		http.Error(w, "Erro ao criar tabela: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Tabela criada!"))
}

// Deleta tabela
func DeleteTable(w http.ResponseWriter, r *http.Request) {
	projectIDStr := mux.Vars(r)["id"]
	table := mux.Vars(r)["table"]
	projectID, _ := strconv.Atoi(projectIDStr)

	var project config.Project
	row := config.MasterDB.QueryRow("SELECT id, name, database_name, api_key FROM projects WHERE id = ?", projectID)
	row.Scan(&project.ID, &project.Name, &project.Database, &project.ApiKey)

	db, err := config.GetDBConnection(&project)
	if err != nil {
		http.Error(w, "Erro ao conectar banco do projeto", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	if err != nil {
		http.Error(w, "Erro ao deletar tabela: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Tabela deletada!"))
}
