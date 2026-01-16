package models

import "time"

// ============================================================================
// ENTITY MODELS - Estruturas de entidades do domínio
// ============================================================================

// Project - Representa um projeto no sistema
type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`         // Prefixo único para tabelas
	Description string    `json:"description"`
	ApiKey      string    `json:"api_key"`
	Type        string    `json:"type"`
	Version     string    `json:"version"`
	Status      string    `json:"status"`       // active, inactive, blocked
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Instance - Representa uma instância de um projeto
type Instance struct {
	ID          int64                  `json:"id"`
	ProjectID   int64                  `json:"project_id"`
	Name        string                 `json:"name"`
	Code        string                 `json:"code"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"` // active, inactive, archived
	Settings    map[string]interface{} `json:"settings"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}