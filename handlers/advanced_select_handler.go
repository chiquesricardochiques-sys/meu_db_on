package handlers

import (
	"encoding/json"
	"net/http"

	"meu-provedor/services/data_service"
	"meu-provedor/models"
)

/*
====================================================
REQUEST BODY – ADVANCED SELECT
====================================================
*/



/*
====================================================
HANDLER
====================================================
*/

func AdvancedSelectHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AdvancedSelectRequest

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
