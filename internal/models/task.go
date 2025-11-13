package models

import "test/internal/dto"

type TaskStatus string

const (
	Queued     TaskStatus = "Queued"
	Processing TaskStatus = "TaskStatus"
	Done       TaskStatus = "Done"
)

type TaskType string

const (
	CheckURL TaskType = "CheckURL"
	LoadPDF  TaskType = "LoadPDF"
)

type Task struct {
	ID         string   `json:"id"`
	LinksID    []string `json:"links_id"`
	TaskType   TaskType
	TaskStatus TaskStatus
	ResultChan chan dto.LinkListResponse
}
