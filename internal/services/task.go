package services

import (
	"errors"
	"fmt"
	"path/filepath"
	"test/internal/app"
	"test/internal/dto"
	"test/internal/models"
	"test/internal/pkg"
	"test/internal/repository"
)

type TaskService struct {
	TaskRepo *repository.TaskRepostiory
	App      *app.App
}

func NewTaskService(tRepo *repository.TaskRepostiory, app *app.App) *TaskService {
	return &TaskService{
		TaskRepo: tRepo,
		App:      app,
	}
}

func (ts *TaskService) SaveTaskService(data []*models.Task) bool {
	return ts.TaskRepo.SaveTask(data)
}

func (ts *TaskService) CreateTaskForLinkService(linkID string, taskType models.TaskType) (dto.LinkListResponse, error) {
	var result dto.LinkListResponse
	resultChan := make(chan dto.LinkListResponse, 1)
	linkIdSlice := []string{linkID}

	task := &models.Task{
		ID:         pkg.GenerateShortID(6),
		LinksID:    linkIdSlice,
		TaskType:   taskType,
		TaskStatus: models.Queued,
		ResultChan: resultChan,
	}

	if ts.App.Draining.Load() {
		taskArray := make([]*models.Task, 0)
		taskArray = append(taskArray, task)

		if !ts.SaveTaskService(taskArray) {
			return result, errors.New("не удалось сохранить вашу задачу")
		}

		fmt.Println("Задача сохранена")
		return result, nil
	}
	//Загружаем таску в канал
	select {
	case *ts.App.TaskChannel <- *task:
		fmt.Printf("Задача %s в обработке!\n", task.ID)
	default:
		fmt.Printf("Задача %s добавлена в буфер!\n", task.ID)
		ts.App.TaskBuffer = append(ts.App.TaskBuffer, task)
	}

	result = <-resultChan

	close(resultChan)

	return result, nil
}

func Worker(id int, tasksChan <-chan models.Task, app *app.App) {
	defer app.WG.Done()
	for task := range tasksChan {
		var result dto.LinkListResponse
		//Меняем статус задачи
		task.TaskStatus = models.Processing

		result.Links = make(map[string]string)
		path := filepath.Join(app.Config.DataDir, "links.json")
		//Обрабатываем
		if task.TaskType == models.CheckURL {
			if processURLCheck(&result, &task, path) {
				repository.UpdateLink(&result)
				//Сохранить и просто отправить в канал строку типо готово
				//плюсы: можно будет использовать для метода с pdf - отправляя ссылку на файл в тот же канал
				//минусы:надо подумать

				//отправлять доступность ссылок и пдф файл по разным каналам
				//плюсы: надо подумать
				//минусы: реализация сложнее, дополнительный канал - затраты

				//TODO:
				//доделать метод обновления ссылки в базе
				//сделать метод проверки URL
				//сделать метод получения нескольких link по id
				//сделать метод формирования pdf
				//сделать хэндлер для обработки запроса на pdf
				//сделать метод получения тасок из базы(после отключения)

			}

		}

		if (len(app.TaskBuffer) != 0) && (!app.Draining.Load()) {
			app.FlushBuffer()
		}

	}
	fmt.Printf("Worker %d stopped\n", id)
}

func processURLCheck(result *dto.LinkListResponse, task *models.Task, path string) bool {
	linkID := task.LinksID[0]
	data, err := repository.GetLinksByID(path, linkID)
	if err != nil {
		fmt.Printf("ERRROR:%v\n", err)
		return false
	}
	for key := range data {
		result.Links[key] = "available"
	}
	result.LinksID = linkID
	task.TaskStatus = models.Done
	task.ResultChan <- *result
	return true
}
