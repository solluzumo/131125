package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"test/internal/app"
	"test/internal/models"
)

type TaskRepostiory struct {
	App      *app.App
	taskPath string
	mu       *sync.Mutex
}

func NewTaskRepository(app *app.App, path string, mu *sync.Mutex) *TaskRepostiory {
	return &TaskRepostiory{
		App:      app,
		taskPath: path,
		mu:       mu,
	}
}

// Функция сохранения задачи
func (tr *TaskRepostiory) SaveTask(data []*models.TaskModel) bool {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	var data_formed []models.TaskModel

	for _, el := range data {
		data_formed = append(data_formed, *el)
	}

	existingData, err := tr.ReadTaskJson()
	if err != nil {
		fmt.Printf("Ошибка при чтении файла:%v\n", err)
		return false
	}

	existingData = append(existingData, data_formed...)

	err = tr.WriteTaskJSON(&existingData)
	if err != nil {
		fmt.Printf("Ошибка при записи файла:%v\n", err)
		return false
	}
	return true
}

// Функция чтения задач
func (tr *TaskRepostiory) ReadTaskJson() ([]models.TaskModel, error) {

	filePath := filepath.Join(tr.taskPath, "tasksQueue.json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(byteValue) == 0 {
		return []models.TaskModel{}, nil
	}

	var data []models.TaskModel
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// Функция сохранения задач
func (tr *TaskRepostiory) WriteTaskJSON(data *[]models.TaskModel) error {

	filePath := filepath.Join(tr.taskPath, "tasksQueue.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	return err
}
