package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/services/data_service"
)

// Estruturas para request
type UpdateRequest struct {
	ProjectID  int64                     `json:"project_id"`
	InstanceID int64                     `json:"id_instancia"`
	Table      string                    `json:"table"`
	Data       map[string]interface{}    `json:"data"`  // campos a atualizar
	Where      map[string]interface{}    `json:"where"` // filtros simples
	WhereRaw   string                    `json:"where_raw,omitempty"` // filtro customizado opcional
}

// UpdateHandler recebe a requisição HTTP e chama o service
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	count, err := data_service.ExecuteUpdate(req)
	if err != nil {
		http.Error(w, "Update failed: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Update completed",
		"count":   count,
	})
}
