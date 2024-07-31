package models

type Configuration struct {
	ID    int    `json:"id"`
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}
