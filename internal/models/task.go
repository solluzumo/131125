package models

type TaskModel struct {
	ID         string   `json:"id"`
	LinksID    []string `json:"links_id"`
	TaskType   string   `json:"task_type"`
	TaskStatus string   `json:"task_status"`
}
