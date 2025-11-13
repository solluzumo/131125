package services

import (
	"fmt"
	"test/internal/app"
	"test/internal/models"
	"test/internal/repository"
	"time"
)

type TaskService struct {
	TaskRepo *repository.TaskRepostiory
}

func NewTaskService(tRepo *repository.TaskRepostiory) *TaskService {
	return &TaskService{
		TaskRepo: tRepo,
	}
}

func (ts *TaskService) SaveTaskService(data []*models.Task) bool {
	return ts.TaskRepo.SaveTask(data)
}

func Worker(id int, tasks <-chan models.Task, app *app.App) {
	defer app.WG.Done()
	for task := range tasks {
		fmt.Printf("Worker %d processing task %s\n", id, task.ID)
		time.Sleep(15 * time.Second)
		fmt.Printf("Worker %d finished task %s\n", id, task.ID)
		if (len(app.TaskBuffer) != 0) && (!app.Draining.Load()) {
			app.FlushBuffer()
		}
	}
	fmt.Printf("Worker %d stopped\n", id)
}
