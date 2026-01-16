package models

// ============================================================================
// REQUEST MODELS - Estruturas de requisição HTTP
// ============================================================================

// DeleteRequest - Requisição para deletar registros
type DeleteRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   string                 `json:"where_raw,omitempty"`
	Mode       string                 `json:"mode,omitempty"` // "hard" ou "soft"
}

// AdvancedSelectRequest - Requisição para SELECT avançado
type AdvancedSelectRequest struct {
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

// Join - Configuração de JOIN
type Join struct {
	Type  string `json:"type"`  // INNER, LEFT, RIGHT
	Table string `json:"table"`
	Alias string `json:"alias,omitempty"`
	On    string `json:"on"`
}

// AdvancedJoinSelectRequest - Requisição para SELECT com múltiplos JOINs
type AdvancedJoinSelectRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Base       JoinBase               `json:"base"`
	Joins      []JoinItem             `json:"joins,omitempty"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   []string               `json:"where_raw,omitempty"`
	GroupBy    string                 `json:"group_by,omitempty"`
	Having     string                 `json:"having,omitempty"`
	OrderBy    string                 `json:"order_by,omitempty"`
	Limit      int                    `json:"limit,omitempty"`
	Offset     int                    `json:"offset,omitempty"`
}

// JoinBase - Tabela base para JOIN
type JoinBase struct {
	Table   string   `json:"table"`
	Alias   string   `json:"alias,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

// JoinItem - Item de JOIN
type JoinItem struct {
	Type    string   `json:"type"`
	Table   string   `json:"table"`
	Alias   string   `json:"alias,omitempty"`
	On      string   `json:"on"`
	Columns []string `json:"columns,omitempty"`
}

// InsertRequest - Requisição para INSERT único
type InsertRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Data       map[string]interface{} `json:"data"`
}

// BatchInsertRequest - Requisição para INSERT em lote
type BatchInsertRequest struct {
	ProjectID  int64                    `json:"project_id"`
	InstanceID int64                    `json:"id_instancia"`
	Table      string                   `json:"table"`
	Data       []map[string]interface{} `json:"data"`
}

// UpdateRequest - Requisição para UPDATE
type UpdateRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Data       map[string]interface{} `json:"data"`
	Where      map[string]interface{} `json:"where,omitempty"`
	WhereRaw   string                 `json:"where_raw,omitempty"`
}

// BatchUpdateRequest - Requisição para UPDATE em lote
type BatchUpdateRequest struct {
	ProjectID  int64               `json:"project_id"`
	InstanceID int64               `json:"id_instancia"`
	Table      string              `json:"table"`
	Updates    []UpdateItem        `json:"updates"`
}

// UpdateItem - Item individual de update em lote
type UpdateItem struct {
	Data  map[string]interface{} `json:"data"`
	Where map[string]interface{} `json:"where"`
}

// AggregateRequest - Requisição para operações de agregação
type AggregateRequest struct {
	ProjectID  int64                  `json:"project_id"`
	InstanceID int64                  `json:"id_instancia"`
	Table      string                 `json:"table"`
	Operation  string                 `json:"operation"` // COUNT, SUM, AVG, MIN, MAX, EXISTS
	Column     string                 `json:"column,omitempty"`
	Where      map[string]interface{} `json:"where,omitempty"`
}

// ============================================================================
// VALIDATION METHODS
// ============================================================================

// Validate - Valida DeleteRequest
func (r *DeleteRequest) Validate() error {
	if r.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if r.InstanceID <= 0 {
		return ErrInvalidInstanceID
	}
	if r.Table == "" {
		return ErrTableRequired
	}
	return nil
}

// Validate - Valida AdvancedSelectRequest
func (r *AdvancedSelectRequest) Validate() error {
	if r.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if r.InstanceID <= 0 {
		return ErrInvalidInstanceID
	}
	if r.Table == "" {
		return ErrTableRequired
	}
	return nil
}

// Validate - Valida BatchInsertRequest
func (r *BatchInsertRequest) Validate() error {
	if r.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if r.InstanceID <= 0 {
		return ErrInvalidInstanceID
	}
	if r.Table == "" {
		return ErrTableRequired
	}
	if len(r.Data) == 0 {
		return ErrNoDataProvided
	}
	return nil
}

// Validate - Valida UpdateRequest
func (r *UpdateRequest) Validate() error {
	if r.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if r.InstanceID <= 0 {
		return ErrInvalidInstanceID
	}
	if r.Table == "" {
		return ErrTableRequired
	}
	if len(r.Data) == 0 {
		return ErrNoDataProvided
	}
	return nil
}

// Validate - Valida AggregateRequest
func (r *AggregateRequest) Validate() error {
	if r.ProjectID <= 0 {
		return ErrInvalidProjectID
	}
	if r.InstanceID <= 0 {
		return ErrInvalidInstanceID
	}
	if r.Table == "" {
		return ErrTableRequired
	}
	if r.Operation == "" {
		return ErrOperationRequired
	}
	return nil
}