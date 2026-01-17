package table

import (
	"errors"
	"regexp"

	"meu-provedor/engine/table"
	"meu-provedor/models"
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
