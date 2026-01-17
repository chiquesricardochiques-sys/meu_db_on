package project

import (
	"strings"

	"meu-provedor/engine/project"
	"meu-provedor/models"
)

func Create(req models.ProjectRequest) error {
	if req.Name == "" || req.Code == "" {
		return models.ErrInvalidProjectData
	}

	req.Code = strings.ToLower(req.Code)

	exists, err := project.ProjectCodeExists(req.Code)
	if err != nil {
		return err
	}
	if exists {
		return models.ErrProjectCodeExists
	}

	return project.InsertProject(req)
}

func List() ([]models.Project, error) {
	return project.ListProjects()
}

func Update(id int64, req models.ProjectRequest) error {
	if id <= 0 {
		return models.ErrProjectNotFound
	}
	return project.UpdateProject(id, req)
}

func Delete(id int64) error {
	if id <= 0 {
		return models.ErrProjectNotFound
	}
	return project.DeleteProject(id)
}
