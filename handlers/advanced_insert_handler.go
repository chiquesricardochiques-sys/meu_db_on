package handlers

import (
	"encoding/json"
	"net/http"
	"meu-provedor/models"
	"meu-provedor/services/data_service"
)

// ============================================================================
// INSERT HANDLERS
// ============================================================================

// InsertHandler processa requisições de INSERT único
func InsertHandler(w http.ResponseWriter, r *http.Request) {
	var req models.InsertRequest
	
	// Decodificar JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	
	// Executar INSERT
	lastID, err := services.ExecuteInsert(req)
	if err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Retornar resultado
	RespondSuccess(w, map[string]interface{}{
		"success": true,
		"message": "Registro inserido com sucesso",
		"id":      lastID,
	})
}

// BatchInsertHandler processa requisições de INSERT em lote
func BatchInsertHandler(w http.ResponseWriter, r *http.Request) {
	var req models.BatchInsertRequest
	
	// Decodificar JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	
	// Executar BATCH INSERT
	count, err := services.ExecuteBatchInsert(req)
	if err != nil {
		RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Retornar resultado
	RespondSuccess(w, map[string]interface{}{
		"success": true,
		"message": "Registros inseridos com sucesso",
		"count":   count,
	})
}
// depurar
func InsertDebugHandler(w http.ResponseWriter, r *http.Request) {
    var req models.InsertRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        RespondError(w, "JSON inválido", http.StatusBadRequest)
        return
    }

    debugResult := services.ExecuteInsertDebug(req)

    status := http.StatusOK
    if !debugResult.Ok {
        status = http.StatusBadRequest
    }

    RespondJSON(w, status, debugResult)
}



