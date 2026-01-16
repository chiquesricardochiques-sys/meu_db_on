package models

type DeleteRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   string                 `json:"where_raw,omitempty"`
	Mode       string                 `json:"mode,omitempty"` // hard | soft
}