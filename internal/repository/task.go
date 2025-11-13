package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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

	path := filepath.Join(ts.App.Config.DataDir, "tasks.json")

	var data_formed []models.Task

	for _, el := range data {
		data_formed = append(data_formed, *el)
	}

	existingData, err := ReadJson(path)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла:%v\n", err)
		return false
	}

	existingData = append(existingData, data_formed...)

	err = writeJSON(path, existingData)
	if err != nil {
		fmt.Printf("Ошибка при записи файла:%v\n", err)
		return false
	}
	fmt.Printf("Было сохранено %d задач", len(data))
	return true
}

func ReadJson(path string) ([]models.Task, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// если файл пустой — возвращаем пустой срез
	if len(byteValue) == 0 {
		return []models.Task{}, nil
	}

	var data []models.Task
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func writeJSON(filename string, data []models.Task) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Конвертируем структуру данных в JSON-строку
	// json.MarshalIndent используется для читаемого форматирования
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	// Записываем JSON-строку в файл
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
