package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/config"
	"meu-provedor/services/data_service"
)

/*
====================================================
STRUCTS PARA REQUISIÇÃO ADVANCED SELECT
====================================================
*/

type JoinConfig struct {
	Type  string `json:"type"`  // INNER, LEFT, RIGHT
	Table string `json:"table"` // Tabela a juntar
	On    string `json:"on"`    // Condição ON: "tabela1.id = tabela2.fk"
	Alias string `json:"alias"` // Apelido da tabela
}

type AdvancedQueryRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Alias      string                 `json:"alias,omitempty"`
	Type       string                 `json:"type"` // simple, join, group_by, advanced
	Select     []string               `json:"columns,omitempty"`
	Joins      []JoinConfig           `json:"joins,omitempty"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   string                 `json:"where_raw,omitempty"`
	GroupBy    string                 `json:"group_by,omitempty"`
	Having     string                 `json:"having,omitempty"`
	OrderBy    string                 `json:"order_by,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
}

/*
====================================================
HANDLER ADVANCED SELECT
====================================================
*/

func AdvancedSelectHandler(w http.ResponseWriter, r *http.Request) {
	var req AdvancedQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	// Chama o service que monta a query e executa
	result, err := data_service.ExecuteSelect(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
