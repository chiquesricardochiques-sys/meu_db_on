package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"database/sql"
	
	"meu-provedor/config"
	"meu-provedor/security"
	"net/http"
)

// Estrutura para requisições genéricas
type RequestData struct {
	Table string                 `json:"table"`     // Nome da tabela para consulta
	Data  map[string]interface{} `json:"data"`      // Dados a serem inseridos ou atualizados
	Query string                 `json:"query,omitempty"` // Condição de WHERE para UPDATE ou DELETE
}

// Função para validar a API Key - Centraliza a verificação da chave
func validateApiKey(w http.ResponseWriter, r *http.Request) (*config.Project, error) {
	apiKey := r.Header.Get("api_key")
	project, err := security.ValidateApiKey(apiKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return nil, err
	}
	return project, nil
}

// Função auxiliar para obter a conexão com o banco de dados do projeto
func getDBConnection(w http.ResponseWriter, project *config.Project) (*sql.DB, error) {
    db, err := config.GetDBConnection(project)
    if err != nil {
        http.Error(w, "Erro ao conectar banco do projeto", http.StatusInternalServerError)
        return nil, err
    }
    return db, nil
}


// Função de inserção genérica
func Insert(w http.ResponseWriter, r *http.Request) {
    project, err := validateApiKey(w, r)
    if err != nil {
        return
    }

    db, err := getDBConnection(w, project)
    if err != nil {
        return
    }

    var req RequestData
    err = json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, "Erro ao ler body", http.StatusBadRequest)
        return
    }

    columns := ""
    placeholders := ""
    values := []interface{}{}
    i := 0
    for k, v := range req.Data {
        if i > 0 {
            columns += ", "
            placeholders += ", "
        }
        columns += k
        placeholders += "?"
        values = append(values, v)
        i++
    }

    sqlQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", req.Table, columns, placeholders)

    stmt, err := db.Prepare(sqlQuery) // Usando PreparedStatement
    if err != nil {
        http.Error(w, "Erro ao preparar consulta", http.StatusInternalServerError)
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(values...) // Exec com prepared statement
    if err != nil {
        http.Error(w, "Erro ao inserir dados", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("✅ Inserido com sucesso!"))
}


// Função GET genérica - Retorna todos os registros da tabela
func Get(w http.ResponseWriter, r *http.Request) {
	project, err := validateApiKey(w, r)
	if err != nil {
		return
	}

	db, err := getDBConnection(w, project)
	if err != nil {
		return
	}

	var req RequestData
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", req.Table))
	if err != nil {
		log.Println("Erro no Select:", err)
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
			log.Println("Erro ao scan:", err)
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

// Função UPDATE genérica - Atualiza registros com base na consulta de WHERE
func Update(w http.ResponseWriter, r *http.Request) {
	project, err := validateApiKey(w, r)
	if err != nil {
		return
	}

	db, err := getDBConnection(w, project)
	if err != nil {
		return
	}

	var req RequestData
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "É necessário fornecer a condição WHERE em 'query'", http.StatusBadRequest)
		return
	}

	setClause := ""
	valArgs := []interface{}{}
	i := 0
	for k, v := range req.Data {
		if i > 0 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = ?", k)
		valArgs = append(valArgs, v)
		i++
	}

	sqlQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s", req.Table, setClause, req.Query)
	_, err = db.Exec(sqlQuery, valArgs...)
	if err != nil {
		http.Error(w, "Erro ao atualizar dados", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Atualizado com sucesso!"))
}

// Função DELETE genérica - Deleta registros com base na consulta de WHERE
func Delete(w http.ResponseWriter, r *http.Request) {
	project, err := validateApiKey(w, r)
	if err != nil {
		return
	}

	db, err := getDBConnection(w, project)
	if err != nil {
		return
	}

	var req RequestData
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "É necessário fornecer a condição WHERE em 'query'", http.StatusBadRequest)
		return
	}

	sqlQuery := fmt.Sprintf("DELETE FROM %s WHERE %s", req.Table, req.Query)
	_, err = db.Exec(sqlQuery)
	if err != nil {
		http.Error(w, "Erro ao deletar dados", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Deletado com sucesso!"))
}

// Função GET com Query String (para consultas mais específicas)
// Função GET com Query String (para consultas mais específicas)
func GetQueryString(w http.ResponseWriter, r *http.Request) {
    apiKey := r.URL.Query().Get("api_key")
    project, err := security.ValidateApiKey(apiKey)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    db, err := config.GetDBConnection(project) // Alteração aqui
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


