package table

import (
	"errors"
	"regexp"

	"meu-provedor/engine/table"
	"meu-provedor/models"

	"fmt"
	"strings"

	"meu-provedor/config"
	
)

var validName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateName(name string) bool {
	return validName.MatchString(name)
}

func Create(projectCode string, req models.CreateTableRequest) (string, error) {
	if !validateName(req.TableName) {
		return "", errors.New("invalid table name")
	}

	for _, col := range req.Columns {
		if !validateName(col.Name) {
			return "", errors.New("invalid column name: " + col.Name)
		}
	}

	for _, idx := range req.Indexes {
		if idx.Name != "" && !validateName(idx.Name) {
			return "", errors.New("invalid index name: " + idx.Name)
		}
		for _, col := range idx.Columns {
			if !validateName(col) {
				return "", errors.New("invalid column in index: " + col)
			}
		}
	}

	return table.CreateTable(projectCode, req)
}

func List(projectCode string) ([]string, error) {
	return table.ListTables(projectCode)
}

func Delete(projectCode, tableName string) error {
	if !validateName(tableName) {
		return errors.New("invalid table name")
	}
	return table.DropTable(projectCode, tableName)
}




// Coluna/Índice compatível
type ColumnRequest = models.ColumnRequest
type IndexRequest = models.IndexRequest

// GET TABLE DETAILS
func GetDetails(projectCode, tableName string) (models.TableDetail, error) {
	fullTable := fmt.Sprintf("%s_%s", projectCode, tableName)

	// pegar colunas
	rows, err := config.MasterDB.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", fullTable))
	if err != nil {
		return models.TableDetail{}, err
	}
	defer rows.Close()

	var columns []models.ColumnDetail
	for rows.Next() {
		var col models.ColumnDetail
		var extra, key, null, defaultVal, typeVal string
		if err := rows.Scan(&col.Name, &typeVal, &null, &key, &defaultVal, &extra); err != nil {
			return models.TableDetail{}, err
		}
		col.Type = typeVal
		col.Nullable = null == "YES"
		col.Key = key
		col.Extra = extra
		col.Default = defaultVal
		columns = append(columns, col)
	}

	// pegar índices
	idxRows, err := config.MasterDB.Query(fmt.Sprintf("SHOW INDEX FROM %s", fullTable))
	if err != nil {
		return models.TableDetail{}, err
	}
	defer idxRows.Close()

	idxMap := map[string]models.IndexDetail{}
	for idxRows.Next() {
		var table, nonUnique, keyName, seq, columnName, collation, cardinality, subPart, packed, nullVal, indexType, comment, indexComment string
		if err := idxRows.Scan(&table, &nonUnique, &keyName, &seq, &columnName, &collation, &cardinality, &subPart, &packed, &nullVal, &indexType, &comment, &indexComment); err != nil {
			continue
		}

		idx, ok := idxMap[keyName]
		if !ok {
			idx = models.IndexDetail{Name: keyName, Type: "INDEX"}
			if nonUnique == "0" {
				idx.Type = "UNIQUE"
			}
		}
		idx.Columns = append(idx.Columns, columnName)
		idxMap[keyName] = idx
	}

	var indexes []models.IndexDetail
	for _, idx := range idxMap {
		indexes = append(indexes, idx)
	}

	return models.TableDetail{
		Name:    tableName,
		Columns: columns,
		Indexes: indexes,
	}, nil
}

// ADD COLUMN
func AddColumn(projectCode, tableName string, col ColumnRequest) error {
	if col.Name == "" || col.Type == "" {
		return errors.New("invalid column")
	}
	fullTable := fmt.Sprintf("%s_%s", projectCode, tableName)
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", fullTable, col.Name, col.Type)
	if !col.Nullable {
		sql += " NOT NULL"
	}
	if col.Unique {
		sql += " UNIQUE"
	}
	_, err := config.MasterDB.Exec(sql)
	return err
}

// MODIFY COLUMN
func ModifyColumn(projectCode, tableName string, col ColumnRequest) error {
	fullTable := fmt.Sprintf("%s_%s", projectCode, tableName)
	sql := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s", fullTable, col.Name, col.Type)
	if !col.Nullable {
		sql += " NOT NULL"
	}
	if col.Unique {
		sql += " UNIQUE"
	}
	_, err := config.MasterDB.Exec(sql)
	return err
}

// DROP COLUMN
func DropColumn(projectCode, tableName, columnName string) error {
	fullTable := fmt.Sprintf("%s_%s", projectCode, tableName)
	_, err := config.MasterDB.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", fullTable, columnName))
	return err
}

// ADD INDEX
func AddIndex(projectCode, tableName string, idx IndexRequest) error {
	if idx.Name == "" || len(idx.Columns) == 0 {
		return errors.New("invalid index")
	}
	fullTable := fmt.Sprintf("%s_%s", projectCode, tableName)
	var sql string
	if strings.ToUpper(idx.Type) == "UNIQUE" {
		sql = fmt.Sprintf("ALTER TABLE %s ADD UNIQUE INDEX %s (%s)", fullTable, idx.Name, strings.Join(idx.Columns, ","))
	} else {
		sql = fmt.Sprintf("ALTER TABLE %s ADD INDEX %s (%s)", fullTable, idx.Name, strings.Join(idx.Columns, ","))
	}
	_, err := config.MasterDB.Exec(sql)
	return err
}

// DROP INDEX
func DropIndex(projectCode, tableName, indexName string) error {
	fullTable := fmt.Sprintf("%s_%s", projectCode, tableName)
	_, err := config.MasterDB.Exec(fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", fullTable, indexName))
	return err
}

