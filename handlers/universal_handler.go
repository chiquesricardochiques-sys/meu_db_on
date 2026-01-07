package handlers

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "meu-provedor/config"
    "meu-provedor/security"
)

// Estrutura para requisições genéricas
type DBRequest struct {
    Action string                 `json:"action"`            // Ação a ser executada: insert, get, update, delete, query, raw_sql
    Table  string                 `json:"table,omitempty"`   // Nome da tabela (opcional em algumas ações)
    Data   map[string]interface{} `json:"data,omitempty"`    // Dados para inserção ou atualização
    Where  string                 `json:"where,omitempty"`   // Condição WHERE (para update e delete)
    SQL    string                 `json:"sql,omitempty"`     // Comandos SQL personalizados
}

// Estrutura de resposta padrão
type JSONResp struct {
    Status  string      `json:"status"`    // Status da operação: "success" ou "error"
    Message string      `json:"message,omitempty"` // Mensagem de resposta
    Data    interface{} `json:"data,omitempty"`    // Dados retornados pela operação
}

// Handler Principal para executar ações no banco
func UniversalHandler(w http.ResponseWriter, r *http.Request) {
    // 1️⃣ Validação da chave de API
    apiKey := r.Header.Get("api_key")
    project, err := security.ValidateApiKey(apiKey) // Valida a chave de API
    if err != nil {
        respond(w, "error", err.Error(), nil) // Se inválido, retorna erro
        return
    }

    // 2️⃣ Conexão com o banco de dados
    db, err := config.GetDBConnection(project) // Obtém a conexão com o banco de dados
    if err != nil {
        respond(w, "error", "Falha ao conectar ao banco do projeto", nil) // Se falhar, retorna erro
        return
    }

    // 3️⃣ Leitura do JSON enviado na requisição
    var req DBRequest
    err = json.NewDecoder(r.Body).Decode(&req) // Decodifica o corpo da requisição
    if err != nil {
        respond(w, "error", "JSON inválido", nil) // Se falhar, retorna erro
        return
    }

    // 4️⃣ Executar a ação baseada na requisição
    switch req.Action {
    case "insert":
        handleInsert(db, req, w)   // Chama a função para inserir dados
    case "get":
        handleGet(db, req, w)      // Chama a função para buscar dados
    case "update":
        handleUpdate(db, req, w)   // Chama a função para atualizar dados
    case "delete":
        handleDelete(db, req, w)   // Chama a função para deletar dados
    case "query":
        handleQuery(db, req, w)    // Chama a função para executar consulta personalizada
    case "raw_sql":
        handleRawSQL(db, req, w)   // Chama a função para executar SQL cru (admin)
    default:
        respond(w, "error", "Ação '"+req.Action+"' inválida", nil) // Ação não reconhecida
    }
}

// Função para inserir dados na tabela
func handleInsert(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" || len(req.Data) == 0 {
        respond(w, "error", "Campos 'table' e 'data' são obrigatórios", nil) // Verifica se os campos obrigatórios estão preenchidos
        return
    }

    columns := ""
    placeholders := ""
    values := []interface{}{}
    i := 0

    // Monta os dados para a inserção
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

    // Executa a query de inserção
    _, err := db.Exec(query, values...)
    if err != nil {
        respond(w, "error", "Erro ao inserir: "+err.Error(), nil) // Se falhar, retorna erro
        return
    }

    respond(w, "success", "Inserido com sucesso", nil) // Resposta de sucesso
}

// Função para pegar dados de uma tabela
func handleGet(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" {
        respond(w, "error", "Campo 'table' é obrigatório", nil) // Valida se a tabela foi fornecida
        return
    }

    query := fmt.Sprintf("SELECT * FROM %s", req.Table)

    rows, err := db.Query(query) // Executa a query para buscar todos os registros
    if err != nil {
        respond(w, "error", "Erro ao buscar dados: "+err.Error(), nil) // Se falhar, retorna erro
        return
    }
    defer rows.Close()

    results := rowsToJSON(rows) // Converte os resultados para JSON
    respond(w, "success", "OK", results) // Retorna os dados em formato JSON
}

// Função para atualizar dados na tabela
func handleUpdate(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" || len(req.Data) == 0 || req.Where == "" {
        respond(w, "error", "Campos 'table', 'data' e 'where' são obrigatórios", nil) // Verifica se os campos obrigatórios estão preenchidos
        return
    }

    setClause := ""
    values := []interface{}{}
    i := 0

    // Monta a cláusula SET para a query de atualização
    for k, v := range req.Data {
        if i > 0 {
            setClause += ", "
        }
        setClause += fmt.Sprintf("%s = ?", k)
        values = append(values, v)
        i++
    }

    query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", req.Table, setClause, req.Where)

    // Executa a query de atualização
    _, err := db.Exec(query, values...)
    if err != nil {
        respond(w, "error", "Erro ao atualizar: "+err.Error(), nil) // Se falhar, retorna erro
        return
    }

    respond(w, "success", "Atualizado com sucesso", nil) // Resposta de sucesso
}

// Função para deletar dados de uma tabela
func handleDelete(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.Table == "" || req.Where == "" {
        respond(w, "error", "Campos 'table' e 'where' são obrigatórios", nil) // Verifica se os campos obrigatórios estão preenchidos
        return
    }

    query := fmt.Sprintf("DELETE FROM %s WHERE %s", req.Table, req.Where)

    // Executa a query de deleção
    _, err := db.Exec(query)
    if err != nil {
        respond(w, "error", "Erro ao deletar: "+err.Error(), nil) // Se falhar, retorna erro
        return
    }

    respond(w, "success", "Deletado com sucesso", nil) // Resposta de sucesso
}

// Função para executar uma consulta SQL personalizada
func handleQuery(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.SQL == "" {
        respond(w, "error", "Campo 'sql' é obrigatório", nil) // Verifica se a query foi fornecida
        return
    }

    rows, err := db.Query(req.SQL) // Executa a query personalizada
    if err != nil {
        respond(w, "error", "Erro ao executar query: "+err.Error(), nil) // Se falhar, retorna erro
        return
    }
    defer rows.Close()

    results := rowsToJSON(rows) // Converte os resultados para JSON
    respond(w, "success", "OK", results) // Retorna os dados em formato JSON
}

// Função para executar comandos SQL diretamente (usado para administração)
func handleRawSQL(db *sql.DB, req DBRequest, w http.ResponseWriter) {
    if req.SQL == "" {
        respond(w, "error", "Campo 'sql' é obrigatório", nil) // Verifica se o SQL foi fornecido
        return
    }

    // Executa o SQL fornecido diretamente no banco
    _, err := db.Exec(req.SQL)
    if err != nil {
        respond(w, "error", "Erro ao executar raw_sql: "+err.Error(), nil) // Se falhar, retorna erro
        return
    }

    respond(w, "success", "Raw SQL executado", nil) // Resposta de sucesso
}

// Função para converter o resultado das queries para JSON
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

// Função para responder com um JSON padrão
func respond(w http.ResponseWriter, status string, msg string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(JSONResp{
        Status:  status,
        Message: msg,
        Data:    data,
    })
}
