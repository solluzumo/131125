package domain

import "context"

type TaskType string

const (
	CheckURL   TaskType = "CheckURL"
	LoadPDF    TaskType = "LoadPDF"
	GiveResult TaskType = "GiveResult"
)

type TaskDomain struct {
	ID         string
	LinkID     []string          //["setID1","setID2"]
	LinksSets  map[string]string //{"link1":"valu1","link2":"value2"...}
	Ctx        context.Context
	TaskType   TaskType
	ResultChan chan interface{}
}
