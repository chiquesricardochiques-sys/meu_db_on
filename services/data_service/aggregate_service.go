package data_service

import (
	"database/sql"
	"errors"

	"meu-provedor/engine/query"
)

type AggregateRequest struct {
	ProjectID  int64
	InstanceID int64
	Table      string
	Operation  string
	Column     string
	Where      map[string]interface{}
}

func ExecuteAggregate(
	db *sql.DB,
	table string,
	req AggregateRequest,
) (interface{}, error) {

	builder := query.NewAggregate(table, "", req.Operation, req.Column)

	// filtro obrigatório
	builder.AddWhere("id_instancia = ?", req.InstanceID)

	for k, v := range req.Where {
		builder.AddWhere(k+" = ?", v)
	}

	sqlQuery := builder.Build()

	var result interface{}
	err := db.QueryRow(sqlQuery, builder.Values...).Scan(&result)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("no result")
	}

	return result, nil
}
