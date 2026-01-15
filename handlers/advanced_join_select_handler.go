package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/services/data_service"
)

/*
====================================================
REQUEST BODY – ADVANCED JOIN SELECT
====================================================
*/

type JoinBase struct {
	Table   string   `json:"table"`
	Alias   string   `json:"alias,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

type JoinItem struct {
	Type    string   `json:"type"`
	Table   string   `json:"table"`
	Alias   string   `json:"alias,omitempty"`
	On      string   `json:"on"`
	Columns []string `json:"columns,omitempty"`
}

type AdvancedJoinSelectRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Base       JoinBase               `json:"base"`
	Joins      []JoinItem             `json:"joins,omitempty"`
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

func AdvancedJoinSelectHandler(w http.ResponseWriter, r *http.Request) {
	var req AdvancedJoinSelectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result, err := data_service.ExecuteAdvancedJoinSelect(req)
	if err != nil {
		http.Error(w, "Query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
