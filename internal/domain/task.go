package domain

type TaskStatus string

const (
	Queued     TaskStatus = "Queued"
	Processing TaskStatus = "TaskStatus"
	Done       TaskStatus = "Done"
	FromMemory TaskStatus = "FromMemory"
)

type TaskType string

const (
	CheckURL   TaskType = "CheckURL"
	LoadPDF    TaskType = "LoadPDF"
	GiveResult TaskType = "GiveResult"
)

type TaskDomain struct {
	ID         string   `json:"id"`
	LinksID    []string `json:"links_id"`
	TaskType   TaskType
	TaskStatus TaskStatus
	ResultChan chan interface{}
}
