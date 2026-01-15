package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/services/data_service"
)

/*
====================================================
REQUEST BODY – ADVANCED SELECT
====================================================
*/

type AdvancedSelectRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Alias      string                 `json:"alias,omitempty"`
	Select     []string               `json:"select,omitempty"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   []string               `json:"where_raw,omitempty"`
	GroupBy    string                 `json:"group_by,omitempty"`
	Having     string                 `json:"having,omitempty"`
	OrderBy    string                 `json:"order_by,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
}

/*
====================================================
HANDLER
====================================================
*/

func AdvancedSelectHandler(w http.ResponseWriter, r *http.Request) {
	var req AdvancedSelectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result, err := data_service.ExecuteAdvancedSelect(req)
	if err != nil {
		http.Error(w, "Select failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
