package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	instanceService "meu-provedor/services/instance"
	"meu-provedor/models"
)

func CreateInstance(w http.ResponseWriter, r *http.Request) {
	var req models.InstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	if err := instanceService.Create(req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("INSTANCE CREATED"))
}

func ListInstances(w http.ResponseWriter, r *http.Request) {
	var projectID *int64

	if pid := r.URL.Query().Get("project_id"); pid != "" {
		id, err := strconv.ParseInt(pid, 10, 64)
		if err != nil {
			http.Error(w, "invalid project_id", 400)
			return
		}
		projectID = &id
	}

	instances, err := instanceService.List(projectID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(instances)
}

func UpdateInstance(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	var req models.InstanceRequest
	json.NewDecoder(r.Body).Decode(&req)

	if err := instanceService.Update(id, req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("INSTANCE UPDATED"))
}

func DeleteInstance(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)

	if err := instanceService.Delete(id); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("INSTANCE DELETED"))
}
