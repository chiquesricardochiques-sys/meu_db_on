package config

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

// Estrutura que representa um projeto cadastrado
type Project struct {
    ID          int
    Name        string
    Database    string
    ApiKey      string
}

// Mapa global de conexões por banco
var Databases = map[string]*sql.DB{}

// Banco master que contém a tabela de projetos
var MasterDB *sql.DB

// Conecta no banco master
// func ConnectMaster() {
//     LoadEnv() // carrega variáveis do .env

//     user := GetEnv("MYSQL_USER")
//     pass := GetEnv("MYSQL_PASS")
//     host := GetEnv("MYSQL_HOST")
//     dbName := GetEnv("MASTER_DB")

//     dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, dbName)
//     var err error
//     MasterDB, err = sql.Open("mysql", dsn)
//     if err != nil {
//         log.Fatalf("Erro ao conectar no banco master: %v", err)
//     }

//     if err := MasterDB.Ping(); err != nil {
//         log.Fatalf("Erro ao pingar banco master: %v", err)
//     }

//     log.Println("✅ Conectado ao banco master:", dbName)
// }
func ConnectMaster() {
    LoadEnv() // carrega variáveis do .env (opcional em Railway)

    user := GetEnv("MYSQLUSER")
    pass := GetEnv("MYSQLPASSWORD")
    host := GetEnv("MYSQLHOST") + ":" + GetEnv("MYSQLPORT")
    dbName := GetEnv("MYSQLDATABASE") // banco master criado no Railway

    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, dbName)
    var err error
    MasterDB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Erro ao conectar no banco master: %v", err)
    }

    if err := MasterDB.Ping(); err != nil {
        log.Fatalf("Erro ao pingar banco master: %v", err)
    }

    log.Println("✅ Conectado ao banco master:", dbName)
}


// Retorna o projeto baseado na API KEY
func GetProjectByApiKey(apiKey string) (*Project, error) {
    var project Project
    query := "SELECT id, name, database_name, api_key FROM projects WHERE api_key = ? LIMIT 1"
    row := MasterDB.QueryRow(query, apiKey)
    err := row.Scan(&project.ID, &project.Name, &project.Database, &project.ApiKey)
    if err != nil {
        return nil, err
    }
    return &project, nil
}

// Retorna a conexão para o banco do projeto
func GetDBConnection(project *Project) (*sql.DB, error) {
    // Se já existe conexão, retorna do cache
    if db, ok := Databases[project.Database]; ok {
        return db, nil
    }

    // Se não existe, cria nova conexão
    user := GetEnv("MYSQL_USER")
    pass := GetEnv("MYSQL_PASS")
    host := GetEnv("MYSQL_HOST")
    dbName := project.Database

    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, dbName)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    Databases[dbName] = db
    log.Println("✅ Conectado ao banco do projeto:", dbName)
    return db, nil
}
