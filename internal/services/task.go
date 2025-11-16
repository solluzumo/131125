package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"test/internal/app"
	"test/internal/domain"
	"test/internal/dto"
	"test/internal/pkg"
	"test/internal/repository"
)

type TaskService struct {
	TaskRepo *repository.TaskRepostiory
	LinkRepo *repository.LinkRepostiory
	App      *app.App
}

func NewTaskService(tRepo *repository.TaskRepostiory, lRepo *repository.LinkRepostiory, app *app.App) *TaskService {
	return &TaskService{
		TaskRepo: tRepo,
		LinkRepo: lRepo,
		App:      app,
	}
}

// Функция создания задачи и её загрузки в канал
func (ts *TaskService) CreateTaskForLinkService(ctx context.Context, task *domain.TaskDomain) (interface{}, error) {

	//Если сервис в процессе завершения работы
	if ts.App.Draining.Load() {
		return nil, errors.New("draining")
	}

	//Если сервис НЕ в процессе завершения работы

	//Загружаем таску в канал
	select {
	case *ts.App.TaskChannel <- *task:
		fmt.Printf("Задача %s в обработке!\n", task.ID)
	case <-ctx.Done():
		return nil, errors.New("request canceled by client")
	default:
		return nil, errors.New("channel is full")
	}

	//Ожидаем результат и возвращаем
	result := <-task.ResultChan

	//Закрываем канал для реузльтатов
	close(task.ResultChan)

	return result, nil
}

// Функция обёртка для проверки URL
func (ts *TaskService) CheckURL(task *domain.TaskDomain) (*dto.LinkListResponse, error) {

	var result dto.LinkListResponse

	result.Links = make(map[string]string)

	if err := ts.ProcessURLCheck(&result, task); err != nil {
		return nil, err
	}

	return &result, nil
}

// Функция обёртка для генерации PDF файла
func (ts *TaskService) GeneratePDF(task *domain.TaskDomain) (string, error) {

	fileName := pkg.PDFNameFromIDs(task.LinkID)

	//Путь до pdf
	pdfPath := filepath.Join(ts.App.Config.PdfDir, fileName)

	//Если файл не существует, начинаем процесс создания, иначе просто возвращаем путь
	if _, err := os.Stat(pdfPath); err != nil {
		if err := ts.ProcessPDF(task, fileName); err != nil {
			return "", err
		}
	}

	return pdfPath, nil
}

// Функция обработки задачи по основным типам
func (ts *TaskService) ProcessTask(task domain.TaskDomain) error {

	switch task.TaskType {

	//Если задача на проверку адресов
	case domain.CheckURL:
		result, err := ts.CheckURL(&task)
		if err != nil {
			return err
		}
		task.ResultChan <- result

	//Если задача на загрузку ПДФ
	case domain.LoadPDF:
		path, err := ts.GeneratePDF(&task)
		if err != nil {
			return err
		}
		task.ResultChan <- path
	}

	return nil
}

// Функция Воркера
func Worker(id int, tasksChan <-chan domain.TaskDomain, taskService *TaskService, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasksChan {
		fmt.Printf("Worker %d WORKING on task %s\n", id, task.ID)
		if err := taskService.ProcessTask(task); err != nil {
			fmt.Printf("Worker %d: ERROR %v\n", id, err)
		}
		fmt.Printf("Worker %d has finished task %s\n", id, task.ID)

	}
	fmt.Printf("Worker %d stopped\n", id)
}

// Функция создания PDF с ссылками
func (ts *TaskService) ProcessPDF(task *domain.TaskDomain, fileName string) error {

	var links []string
	//получаем ["link1-valu1","link2-value2",...] для всех наборов ссылок по их linkID
	for _, linkID := range task.LinkID {

		linkData, err := ts.LinkRepo.GetLinksByID(linkID)
		if err != nil {
			return err
		}

		for url, status := range linkData {
			links = append(links, fmt.Sprintf("%s-%s", url, status))
		}
	}

	//сохраняем в pdf
	err := ts.LinkRepo.SavePDF(links, fileName)
	if err != nil {
		return err
	}

	return nil
}

// Функция проверки доступности ссылок
func (ts *TaskService) ProcessURLCheck(result *dto.LinkListResponse, task *domain.TaskDomain) error {
	linkID := task.LinkID[0]

	data := task.LinksSets

	for URL := range data {
		result.Links[URL] = pkg.SendRequest(URL)
	}
	result.LinksID = linkID

	ts.LinkRepo.SaveLinks(result.Links, result.LinksID)

	return nil
}
