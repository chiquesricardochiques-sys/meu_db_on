package config

import (
    "crypto/tls"
    "crypto/x509"
    "database/sql"
    "io/ioutil"
    "log"
    "os"

    "github.com/go-sql-driver/mysql"
)

// Estrutura que representa um projeto cadastrado
type Project struct {
    ID     int
    Name   string
    Prefix string // Prefixo único para as tabelas do projeto
    ApiKey string
}

// Mapa global de conexões (apenas 1 database agora)
var Databases = map[string]*sql.DB{}

// Banco master que contém a tabela de projetos
var MasterDB *sql.DB

// Conecta no banco master (Aiven SSL)
func ConnectMaster() {
    user := os.Getenv("MYSQLUSER")
    pass := os.Getenv("MYSQLPASSWORD")
    host := os.Getenv("MYSQLHOST")
    port := os.Getenv("MYSQLPORT")
    dbName := os.Getenv("MYSQLDATABASE")

    // Carrega certificado CA
    rootCertPool := x509.NewCertPool()
    pem, err := ioutil.ReadFile("ca.pem")
    if err != nil {
        log.Fatalf("Erro ao ler CA: %v", err)
    }
    if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
        log.Fatalf("Erro ao adicionar CA")
    }

    // Registra TLS config
    tlsConfig := &tls.Config{
        RootCAs: rootCertPool,
    }
    err = mysql.RegisterTLSConfig("aiven", tlsConfig)
    if err != nil {
        log.Fatalf("Erro ao registrar TLS config: %v", err)
    }

    dsn := user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + dbName + "?parseTime=true&tls=aiven"

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Erro ao conectar no MySQL Aiven: %v", err)
    }

    if err := db.Ping(); err != nil {
        log.Fatalf("Erro ao pingar banco: %v", err)
    }

    MasterDB = db
    log.Println("✅ Conectado ao banco master Aiven:", dbName)
}

// Retorna o projeto baseado na API KEY
func GetProjectByApiKey(apiKey string) (*Project, error) {
    var project Project
    query := "SELECT id, name, prefix, api_key FROM projects WHERE api_key = ? LIMIT 1"
    row := MasterDB.QueryRow(query, apiKey)
    err := row.Scan(&project.ID, &project.Name, &project.Prefix, &project.ApiKey)
    if err != nil {
        return nil, err
    }
    return &project, nil
}

// Retorna a conexão para o único database (Aiven SSL)
func GetDBConnection(project *Project) (*sql.DB, error) {
    dbName := os.Getenv("MYSQLDATABASE") // único database agora

    if db, ok := Databases[dbName]; ok {
        return db, nil
    }

    user := os.Getenv("MYSQLUSER")
    pass := os.Getenv("MYSQLPASSWORD")
    host := os.Getenv("MYSQLHOST") + ":" + os.Getenv("MYSQLPORT")

    dsn := user + ":" + pass + "@tcp(" + host + ")/" + dbName + "?parseTime=true&tls=aiven"

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    Databases[dbName] = db
    log.Println("✅ Conectado ao único banco:", dbName)
    return db, nil
}
