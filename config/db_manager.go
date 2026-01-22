package config

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-sql-driver/mysql"
)

// ============================================================================
// DATABASE CONNECTION
// ============================================================================

// MasterDB é a conexão global com o banco de dados
var MasterDB *sql.DB

// ConnectMaster estabelece conexão com o banco master
func ConnectMaster() error {
	user := os.Getenv("MYSQLUSER")
	pass := os.Getenv("MYSQLPASSWORD")
	host := os.Getenv("MYSQLHOST")
	port := os.Getenv("MYSQLPORT")
	dbName := os.Getenv("MYSQLDATABASE")

	if user == "" || pass == "" || host == "" || port == "" || dbName == "" {
		return fmt.Errorf("variáveis de ambiente do banco não configuradas")
	}

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

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("erro ao conectar no MySQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("erro ao pingar banco: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	MasterDB = db
	log.Println("✅ Conectado ao banco master:", dbName)

	return nil
}

// CloseDB fecha a conexão com o banco
func CloseDB() error {
	if MasterDB != nil {
		return MasterDB.Close()
	}
	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GetProjectByID retorna projeto pelo ID
func GetProjectByID(projectID int) (*Project, error) {
	var project Project
	query := "SELECT id, name, code, api_key FROM projects WHERE id = ? LIMIT 1"
	row := MasterDB.QueryRow(query, projectID)
	err := row.Scan(&project.ID, &project.Name, &project.Code, &project.ApiKey)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProjectByApiKey retorna projeto pelo apiKey
func GetProjectByApiKey(apiKey string) (*Project, error) {
	var project Project
	row := MasterDB.QueryRow("SELECT id, name, code, api_key FROM projects WHERE api_key=? LIMIT 1", apiKey)
	err := row.Scan(&project.ID, &project.Name, &project.Code, &project.ApiKey)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProjectCodeByID retorna o code de um projeto dado seu ID
func GetProjectCodeByID(projectID int) (string, error) {
	var code string
	query := "SELECT code FROM projects WHERE id = ? LIMIT 1"
	row := MasterDB.QueryRow(query, projectID)
	err := row.Scan(&code)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar code do projeto: %w", err)
	}
	return code, nil
}

// BuildTableName constrói o nome completo da tabela com prefixo do projeto
func BuildTableName(project *Project, table string) string {
	return fmt.Sprintf("%s_%s", project.Code, table)
}

// RowsToMap converte sql.Rows para []map[string]interface{}
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
			val := columns[i]
			if b, ok := val.([]byte); ok {
				m[colName] = string(b)
			} else {
				m[colName] = val
			}
		}
		results = append(results, m)
	}
	return results, nil
}

// ============================================================================
// PROJECT STRUCT
// ============================================================================

type Project struct {
	ID     int
	Name   string
	Code   string
	ApiKey string
}


