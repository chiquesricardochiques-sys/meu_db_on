package models


type ProjectRequest struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	ApiKey  string `json:"api_key"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Status  string `json:"status"`
}


