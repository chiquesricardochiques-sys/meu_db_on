package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"meu-provedor/models"
	"meu-provedor/services/data_service"
)

// DeleteHandler processa a requisição de delete com modo hard ou soft
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	var req models.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mode := strings.ToLower(req.Mode)
	if mode == "" {
		mode = "hard"
	}

	var (
		count int64
		err   error
	)

	switch mode {
	case "soft":
		count, err = data_service.ExecuteSoftDelete(req)
	default:
		count, err = data_service.ExecuteHardDelete(req)
	}

	if err != nil {
		http.Error(w, "Delete failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"mode":    mode,
		"count":   count,
	})
}
