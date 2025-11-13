package models

type Task struct {
	ID   string   `json:"id"`
	Data []string `json:"data"`
}
