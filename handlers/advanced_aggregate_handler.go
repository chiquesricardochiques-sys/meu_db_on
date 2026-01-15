package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/config"
	"meu-provedor/services/data_service"
)

type AggregateHTTPRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Operation  string                 `json:"operation"`
	Column     string                 `json:"column,omitempty"`
	Where      map[string]interface{} `json:"where,omitempty"`
}

func AdvancedAggregate(w http.ResponseWriter, r *http.Request) {
	var req AggregateHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	projectCode, err := getProjectCodeByID(req.ProjectID)
	if err != nil {
		http.Error(w, "project not found", 404)
		return
	}

	table, err := buildTableName(projectCode, req.Table)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	result, err := data_service.ExecuteAggregate(
		config.MasterDB,
		table,
		data_service.AggregateRequest{
			ProjectID:  req.ProjectID,
			InstanceID: req.InstanceID,
			Table:      table,
			Operation:  req.Operation,
			Column:     req.Column,
			Where:      req.Where,
		},
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
	})
}
