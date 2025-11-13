package repository

import (
	"fmt"
	"path/filepath"
	"test/internal/app"
	"test/internal/models"
)

type TaskRepostiory struct {
	App *app.App
}

func NewTaskRepository(app *app.App) *TaskRepostiory {
	return &TaskRepostiory{
		App: app,
	}
}

func (ts *TaskRepostiory) SaveTask(data []*models.Task) bool {

	path := filepath.Join(ts.App.Config.DataDir, "tasksQueue.json")

	var data_formed []models.Task

	for _, el := range data {
		data_formed = append(data_formed, *el)
	}

	existingData, err := ReadJson[models.Task](path)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла:%v\n", err)
		return false
	}

	existingData = append(existingData, data_formed...)

	err = WriteJSON(path, existingData)
	if err != nil {
		fmt.Printf("Ошибка при записи файла:%v\n", err)
		return false
	}
	fmt.Printf("Было сохранено %d задач", len(data))
	return true
}
