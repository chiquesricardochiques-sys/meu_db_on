package services

import (
	"database/sql"
	"fmt"
	"time"

	"meu-provedor/config"
	"meu-provedor/engine/query"
)

// ExecuteSoftDelete executa soft delete (UPDATE deleted_at)
func ExecuteSoftDelete(req DeleteRequest) (int64, error) {
	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		return 0, err
	}

	table, err := buildTableName(projectCode, req.Table)
	if err != nil {
		return 0, err
	}

	// garante coluna deleted_at
	if err := ensureSoftDeleteColumn(config.MasterDB, table); err != nil {
		return 0, err
	}

	builder := query.NewSoftDelete(table)

	builder.AddWhere("id_instancia = ?", req.InstanceID)

	for k, v := range req.Where {
		builder.AddWhere(fmt.Sprintf("%s = ?", k), v)
	}

	if req.WhereRaw != "" {
		builder.AddRawWhere(req.WhereRaw)
	}

	sqlQuery, args := builder.Build(time.Now())

	result, err := config.MasterDB.Exec(sqlQuery, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// cria coluna deleted_at apenas se não existir
func ensureSoftDeleteColumn(db *sql.DB, table string) error {
	var exists int
	queryCheck := `
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = 'deleted_at'
	`

	if err := db.QueryRow(queryCheck, table).Scan(&exists); err != nil {
		return err
	}

	if exists == 0 {
		alter := fmt.Sprintf("ALTER TABLE %s ADD COLUMN deleted_at DATETIME NULL", table)
		_, err := db.Exec(alter)
		return err
	}

	return nil
}

