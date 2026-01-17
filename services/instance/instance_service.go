package instance

import (
	"errors"

	"meu-provedor/engine/instance"
	"meu-provedor/models"
)

func Create(req models.InstanceRequest) error {
	if req.ProjectID <= 0 {
		return errors.New("project_id inválido")
	}
	if req.Name == "" || req.Code == "" {
		return errors.New("name e code são obrigatórios")
	}

	return instance.InsertInstance(req)
}

func List(projectID *int64) ([]models.Instance, error) {
	return instance.ListInstances(projectID)
}

func Update(id int64, req models.InstanceRequest) error {
	if id <= 0 {
		return errors.New("id inválido")
	}
	return instance.UpdateInstance(id, req)
}

func Delete(id int64) error {
	if id <= 0 {
		return errors.New("id inválido")
	}
	return instance.DeleteInstance(id)
}
