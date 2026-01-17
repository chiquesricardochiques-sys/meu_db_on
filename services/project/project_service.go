package project
import (
	
	
	"meu-provedor/engine/project"
	"meu-provedor/models"
)
func Create(req models.ProjectRequest) error {
    // validação de negócio
    if req.Code == "" {
        return errors.New("code obrigatório")
    }

    // regra: code é único?
    exists, _ :=  project.ProjectCodeExists(req.Code)
    if exists {
        return errors.New("project code já existe")
    }

    return project.InsertProject(req)
}

