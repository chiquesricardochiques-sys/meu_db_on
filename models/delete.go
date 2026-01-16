package models

type DeleteRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   string                 `json:"where_raw,omitempty"`
	Mode       string                 `json:"mode,omitempty"` // hard | soft
}


type AdvancedSelectRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Alias      string                 `json:"alias,omitempty"`
	Select     []string               `json:"select,omitempty"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   []string               `json:"where_raw,omitempty"`
	GroupBy    string                 `json:"group_by,omitempty"`
	Having     string                 `json:"having,omitempty"`
	OrderBy    string                 `json:"order_by,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
}

type BatchInsertRequest struct {
    ProjectID  int64                    `json:"project_id"`
    InstanceID int64                    `json:"id_instancia"`
    Table      string                   `json:"table"`
    Data       []map[string]interface{} `json:"data"`
}



type Join struct {
    Type  string `json:"type"`
    Table string `json:"table"`
    Alias string `json:"alias"`
    On    string `json:"on"`
}

type AdvancedQueryRequest struct {
    ProjectID  int64                  `json:"project_id"`
    InstanceID int64                  `json:"id_instancia"`
    Table      string                 `json:"table"`
    Alias      string                 `json:"alias,omitempty"`
    Select     []string               `json:"select,omitempty"`
    Joins      []Join                 `json:"joins,omitempty"`
    Where      map[string]interface{} `json:"where,omitempty"`
    WhereRaw   string                 `json:"where_raw,omitempty"`
    GroupBy    string                 `json:"group_by,omitempty"`
    Having     string                 `json:"having,omitempty"`
    OrderBy    string                 `json:"order_by,omitempty"`
    Limit      int                    `json:"limit,omitempty"`
    Offset     int                    `json:"offset,omitempty"`
}


type UpdateRequest struct {
    ProjectID  int64                  `json:"project_id"`
    InstanceID int64                  `json:"id_instancia"`
    Table      string                 `json:"table"`
    Data       map[string]interface{} `json:"data"`             // campos a atualizar
    Where      map[string]interface{} `json:"where,omitempty"`  // filtros simples
    WhereRaw   string                 `json:"where_raw,omitempty"`
}

