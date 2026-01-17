// services/project/project_service.go
package project

func Create(req models.ProjectRequest) error {
    // validação de negócio
    if req.Code == "" {
        return errors.New("code obrigatório")
    }

    // regra: code é único?
    exists, _ := engine.ProjectCodeExists(req.Code)
    if exists {
        return errors.New("project code já existe")
    }

    return engine.InsertProject(req)
}
