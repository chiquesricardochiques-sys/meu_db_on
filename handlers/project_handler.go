package handlers

import (
	"encoding/json"
	"fmt"
	"meu-provedor/config"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
)

// Estrutura para criar/atualizar projeto
type ProjectRequest struct {
	Name   string `json:"name"`
	ApiKey string `json:"api_key"`
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

	sqlQuery := "INSERT INTO projects (name, api_key) VALUES (?, ?)"
	_, err := config.MasterDB.Exec(sqlQuery, req.Name, req.ApiKey)
	if err != nil {
		http.Error(w, "Erro ao criar projeto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Projeto criado com sucesso!"))
}

// Lista todos os projetos
func ListProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := config.MasterDB.Query("SELECT id, name, api_key FROM projects")
	if err != nil {
		http.Error(w, "Erro ao buscar projetos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	projects := []config.Project{}
	for rows.Next() {
		var p config.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.ApiKey); err != nil {
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

	sqlQuery := "UPDATE projects SET name = ?, api_key = ? WHERE id = ?"
	_, err := config.MasterDB.Exec(sqlQuery, req.Name, req.ApiKey, id)
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

	// Remove projeto do banco
	_, err := config.MasterDB.Exec("DELETE FROM projects WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Erro ao deletar projeto: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Projeto deletado!"))
}

// ===================== TABELAS =====================

// Função para validar nomes de tabela/coluna (apenas letras, números e _)
func isValidName(name string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return re.MatchString(name)
}

// Cria nova tabela no projeto
func CreateTable(w http.ResponseWriter, r *http.Request) {
	projectIDStr := mux.Vars(r)["id"]
	projectID, _ := strconv.Atoi(projectIDStr)

	// Busca projeto
	var project config.Project
	row := config.MasterDB.QueryRow("SELECT id, name, api_key FROM projects WHERE id = ?", projectID)
	if err := row.Scan(&project.ID, &project.Name, &project.ApiKey); err != nil {
		http.Error(w, "Projeto não encontrado", http.StatusNotFound)
		return
	}

	var req TableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	if !isValidName(req.TableName) {
		http.Error(w, "Nome da tabela inválido", http.StatusBadRequest)
		return
	}

	columns := ""
	i := 0
	for col, tipo := range req.Columns {
		if !isValidName(col) {
			http.Error(w, "Nome de coluna inválido: "+col, http.StatusBadRequest)
			return
		}
		if i > 0 {
			columns += ", "
		}
		columns += fmt.Sprintf("%s %s", col, tipo)
		i++
	}

	tableName := fmt.Sprintf("projeto_%d_%s", project.ID, req.TableName)
	sqlQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, columns)

	_, err := config.MasterDB.Exec(sqlQuery)
	if err != nil {
		http.Error(w, "Erro ao criar tabela: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Tabela criada!"))
}

// Deleta tabela do projeto
func DeleteTable(w http.ResponseWriter, r *http.Request) {
	projectIDStr := mux.Vars(r)["id"]
	table := mux.Vars(r)["table"]
	projectID, _ := strconv.Atoi(projectIDStr)

	var project config.Project
	row := config.MasterDB.QueryRow("SELECT id, name, api_key FROM projects WHERE id = ?", projectID)
	if err := row.Scan(&project.ID, &project.Name, &project.ApiKey); err != nil {
		http.Error(w, "Projeto não encontrado", http.StatusNotFound)
		return
	}

	if !isValidName(table) {
		http.Error(w, "Nome da tabela inválido", http.StatusBadRequest)
		return
	}

	tableName := fmt.Sprintf("projeto_%d_%s", project.ID, table)
	_, err := config.MasterDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		http.Error(w, "Erro ao deletar tabela: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Tabela deletada!"))
}

