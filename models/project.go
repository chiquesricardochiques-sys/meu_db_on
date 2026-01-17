package models

import "time"

type Project struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	ApiKey    string    `json:"api_key"`
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectRequest struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	ApiKey  string `json:"api_key"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"`
}
