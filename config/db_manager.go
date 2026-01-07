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
    // Pega variáveis de ambiente
    user := os.Getenv("MYSQLUSER")
    pass := os.Getenv("MYSQLPASSWORD")
    host := os.Getenv("MYSQLHOST")
    port := os.Getenv("MYSQLPORT")
    dbName := os.Getenv("MYSQLDATABASE")

    // Carrega certificado CA (baixar do Aiven e colocar na raiz do projeto)
    rootCertPool := x509.NewCertPool()
    pem, err := ioutil.ReadFile("ca.pem") // O arquivo do Aiven
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

    // DSN com TLS
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
    // user := GetEnv("MYSQL_USER")
    // pass := GetEnv("MYSQL_PASS")
    // host := GetEnv("MYSQL_HOST")
    user := GetEnv("MYSQLUSER")
    pass := GetEnv("MYSQLPASSWORD")
    host := os.Getenv("MYSQLHOST") + ":" + os.Getenv("MYSQLPORT")

    dbName := project.Database

    dsn := user + ":" + pass + "@tcp(" + host + ")/" + dbName + "?parseTime=true&tls=aiven"
)
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


