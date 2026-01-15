package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"meu-provedor/services/data_service"
)

// DeleteRequest representa o body esperado
type DeleteRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   string                 `json:"where_raw,omitempty"`
	Mode       string                 `json:"mode,omitempty"` // hard | soft
}

// DeleteHandler processa a requisição
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
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
