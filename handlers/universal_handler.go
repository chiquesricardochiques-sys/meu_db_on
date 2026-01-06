package handlers

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "meu-provedor/config"
    "meu-provedor/security"
)

// Estrutura universal de requisição
type DBRequest struct {
    Action string                 `json:"action"`            // insert, get, update, delete, query, raw_sql
    Table  string                 `json:"table,omitempty"`   // tabela opcional
    Data   map[string]interface{} `json:"data,omitempty"`    // dados para insert/update
    Where  string                 `json:"where,omitempty"`   // condição para update/delete
    SQL    string                 `json:"sql,omitempty"`     // comandos SQL personalizados
}

// Resposta padrão
type JSONResp struct {
    Status  string      `json:"status"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// ----------------------------------
// HANDLER PRINCIPAL
// ----------------------------------
func UniversalHandler(w http.ResponseWriter, r *http.Request) {

    // 1️⃣ Validar API KEY
    apiKey := r.Header.Get("api_key")
    project, err := security.ValidateApiKey(apiKey)
    if err != nil {
        respond(w, "error", err.Error(), nil)
        return
    }

    // 2️⃣ Obter conexão ao banco do projeto
    db, err := config.GetDBConnection(project)
    if err != nil {
        respond(w, "error", "Falha ao conectar ao banco do projeto", nil)
        return
    }

    // 3️⃣ Ler o JSON enviado
    var req DBRequest
    err = json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        respond(w, "error", "JSON inválido", nil)
        return
    }

    // 4️⃣ Executar a ação
    switch req.Action {

    case "insert":
        handleInsert(db, req, w)

    case "get":
        handleGet(db, req, w)

    case "update":
        handleUpdate(db, req, w)

    case "delete":
        handleDelete(db, req, w)

    case "query":
        handleQuery(db, req, w)

    case "raw_sql": // opcional
        handleRawSQL(db, req, w)

    default:
        respond(w, "error", "Ação '"+req.Action+"' inválida", nil)
    }
}

// ----------------------------------
// INSERT
// ----------------------------------
func handleInsert(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" || len(req.Data) == 0 {
        respond(w, "error", "Campos 'table' e 'data' são obrigatórios", nil)
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

    query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", req.Table, columns, placeholders)

    _, err := db.Exec(query, values...)
    if err != nil {
        respond(w, "error", "Erro ao inserir: "+err.Error(), nil)
        return
    }

    respond(w, "success", "Inserido com sucesso", nil)
}

// ----------------------------------
// GET
// ----------------------------------
func handleGet(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" {
        respond(w, "error", "Campo 'table' é obrigatório", nil)
        return
    }

    query := fmt.Sprintf("SELECT * FROM %s", req.Table)

    rows, err := db.Query(query)
    if err != nil {
        respond(w, "error", "Erro ao buscar dados: "+err.Error(), nil)
        return
    }
    defer rows.Close()

    results := rowsToJSON(rows)
    respond(w, "success", "OK", results)
}

// ----------------------------------
// UPDATE
// ----------------------------------
func handleUpdate(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" || len(req.Data) == 0 || req.Where == "" {
        respond(w, "error", "Campos 'table', 'data' e 'where' são obrigatórios", nil)
        return
    }

    setClause := ""
    values := []interface{}{}
    i := 0

    for k, v := range req.Data {
        if i > 0 {
            setClause += ", "
        }
        setClause += fmt.Sprintf("%s = ?", k)
        values = append(values, v)
        i++
    }

    query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", req.Table, setClause, req.Where)

    _, err := db.Exec(query, values...)
    if err != nil {
        respond(w, "error", "Erro ao atualizar: "+err.Error(), nil)
        return
    }

    respond(w, "success", "Atualizado com sucesso", nil)
}

// ----------------------------------
// DELETE
// ----------------------------------
func handleDelete(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" || req.Where == "" {
        respond(w, "error", "Campos 'table' e 'where' são obrigatórios", nil)
        return
    }

    query := fmt.Sprintf("DELETE FROM %s WHERE %s", req.Table, req.Where)

    _, err := db.Exec(query)
    if err != nil {
        respond(w, "error", "Erro ao deletar: "+err.Error(), nil)
        return
    }

    respond(w, "success", "Deletado com sucesso", nil)
}

// ----------------------------------
// QUERY (SELECT personalizado)
// ----------------------------------
func handleQuery(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.SQL == "" {
        respond(w, "error", "Campo 'sql' é obrigatório", nil)
        return
    }

    rows, err := db.Query(req.SQL)
    if err != nil {
        respond(w, "error", "Erro ao executar query: "+err.Error(), nil)
        return
    }
    defer rows.Close()

    results := rowsToJSON(rows)
    respond(w, "success", "OK", results)
}

// ----------------------------------
// RAW SQL - cuidado! (admin)
// ----------------------------------
func handleRawSQL(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.SQL == "" {
        respond(w, "error", "Campo 'sql' é obrigatório", nil)
        return
    }

    _, err := db.Exec(req.SQL)
    if err != nil {
        respond(w, "error", "Erro ao executar raw_sql: "+err.Error(), nil)
        return
    }

    respond(w, "success", "Raw SQL executado", nil)
}

// ----------------------------------
// Utility: converter rows para JSON
// ----------------------------------
func rowsToJSON(rows *sql.Rows) []map[string]interface{} {
    cols, _ := rows.Columns()
    results := []map[string]interface{}{}

    for rows.Next() {
        columns := make([]interface{}, len(cols))
        pointers := make([]interface{}, len(cols))

        for i := range columns {
            pointers[i] = &columns[i]
        }

        rows.Scan(pointers...)

        m := map[string]interface{}{}
        for i, colName := range cols {
            val := pointers[i].(*interface{})
            m[colName] = *val
        }

        results = append(results, m)
    }
    return results
}

// ----------------------------------
// Utility: resposta JSON padrão
// ----------------------------------
func respond(w http.ResponseWriter, status string, msg string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(JSONResp{
        Status:  status,
        Message: msg,
        Data:    data,
    })
}
