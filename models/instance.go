package models

import "time"

type Instance struct {
	ID          int64                  `json:"id"`
	ProjectID   int64                  `json:"project_id"`
	Name        string                 `json:"name"`
	Code        string                 `json:"code"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Settings    map[string]interface{} `json:"settings"`
	CreatedAt   time.Time              `json:"created_at"`
}



type InstanceRequest struct {
	ProjectID   int64                  `json:"project_id"`
	Name        string                 `json:"name"`
	Code        string                 `json:"code"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Settings    map[string]interface{} `json:"settings"`
}
