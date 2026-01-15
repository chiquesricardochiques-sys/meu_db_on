package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/services/data_service"
)

type BatchInsertRequest struct {
	ProjectID  int64                    `json:"project_id"`
	InstanceID int64                    `json:"instance_id"`
	Table      string                   `json:"table"`
	Data       []map[string]interface{} `json:"data"`
}

func BatchInsertHandler(w http.ResponseWriter, r *http.Request) {
	var req BatchInsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if len(req.Data) == 0 {
		http.Error(w, "No data provided", 400)
		return
	}

	count, err := data_service.ExecuteBatchInsert(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Batch insert completed",
		"count":   count,
	})
}
