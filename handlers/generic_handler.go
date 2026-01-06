package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"meu-provedor/config"
	"meu-provedor/security"
	"net/http"
)

// Estrutura para requisições genéricas
type RequestData struct {
	Table string                 `json:"table"`
	Data  map[string]interface{} `json:"data"`
	Query string                 `json:"query,omitempty"` // para GET ou DELETE customizado
}

// Inserir dados genéricos
func Insert(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("api_key")
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

	var req RequestData
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	columns := ""
	values := ""
	valArgs := []interface{}{}
	i := 0
	for k, v := range req.Data {
		if i > 0 {
			columns += ", "
			values += ", "
		}
		columns += k
		values += "?"
		valArgs = append(valArgs, v)
		i++
	}

	sqlQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", req.Table, columns, values)
	_, err = db.Exec(sqlQuery, valArgs...)
	if err != nil {
		log.Println("Erro no Insert:", err)
		http.Error(w, "Erro ao inserir dados", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("✅ Inserido com sucesso!"))
}

// GET genérico simples (retorna todos os registros da tabela)
func Get(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("api_key")
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
// Update genérico
func Update(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("api_key")
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

	var req RequestData
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Erro ao ler body", http.StatusBadRequest)
		return
	}

	// Espera que req.Data tenha os campos a atualizar
	// e req.Query tenha a condição WHERE
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

// Delete genérico
func Delete(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("api_key")
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
