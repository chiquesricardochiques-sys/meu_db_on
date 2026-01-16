package config

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

// ============================================================================
// DATABASE CONNECTION
// ============================================================================

// MasterDB é a conexão global com o banco de dados
var MasterDB *sql.DB

// Estrutura de projeto
type Project struct {
	ID     int
	Name   string
	Prefix string
	ApiKey string
}


// ConnectMaster estabelece conexão com o banco master
func ConnectMaster() error {
	user := os.Getenv("MYSQLUSER")
	pass := os.Getenv("MYSQLPASSWORD")
	host := os.Getenv("MYSQLHOST")
	port := os.Getenv("MYSQLPORT")
	dbName := os.Getenv("MYSQLDATABASE")

	// Validar variáveis obrigatórias
	if user == "" || pass == "" || host == "" || port == "" || dbName == "" {
		return fmt.Errorf("variáveis de ambiente do banco não configuradas")
	}

	// Configurar TLS (se ca.pem existir)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbName)

	if _, err := os.Stat("ca.pem"); err == nil {
		rootCertPool := x509.NewCertPool()
		pem, err := ioutil.ReadFile("ca.pem")
		if err != nil {
			return fmt.Errorf("erro ao ler ca.pem: %w", err)
		}

		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			return fmt.Errorf("erro ao adicionar certificado CA")
		}

		tlsConfig := &tls.Config{
			RootCAs: rootCertPool,
		}

		if err := mysql.RegisterTLSConfig("aiven", tlsConfig); err != nil {
			return fmt.Errorf("erro ao registrar TLS config: %w", err)
		}

		dsn += "&tls=aiven"
	}

	// Conectar ao banco
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("erro ao conectar no MySQL: %w", err)
	}

	// Verificar conexão
	if err := db.Ping(); err != nil {
		return fmt.Errorf("erro ao pingar banco: %w", err)
	}

	// Configurações de pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	MasterDB = db
	log.Println("✅ Conectado ao banco master:", dbName)

	return nil
}

// GetProjectByApiKey retorna projeto pelo apiKey
func GetProjectByApiKey(apiKey string) (*Project, error) {
	var project Project
	row := MasterDB.QueryRow("SELECT id, name, prefix, api_key FROM projects WHERE api_key=? LIMIT 1", apiKey)
	err := row.Scan(&project.ID, &project.Name, &project.Prefix, &project.ApiKey)
	if err != nil {
		return nil, err
	}
	return &project, nil
}


// GetDBConnection retorna a conexão para um projeto
func GetDBConnection(project *Project) (*sql.DB, error) {
	if MasterDB == nil {
		return nil, fmt.Errorf("MasterDB não inicializado")
	}
	return MasterDB, nil
}

// Funções utilitárias usadas pelos services
func BuildTableName(project *Project, table string) string {
	return fmt.Sprintf("%s_%s", project.Prefix, table)
}

func RowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		for i, colName := range cols {
			m[colName] = columns[i]
		}
		results = append(results, m)
	}
	return results, nil
}

// CloseDB fecha a conexão com o banco
func CloseDB() error {
	if MasterDB != nil {
		return MasterDB.Close()
	}
	return nil

}
