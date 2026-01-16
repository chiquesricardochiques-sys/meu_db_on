package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/services/data_service"
	"meu-provedor/models"
)



// UpdateHandler recebe a requisição HTTP e chama o service
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateRequest
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
