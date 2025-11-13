package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"test/internal/app"
	"test/internal/interfaces/httpHandlers"
	"test/internal/models"
	"test/internal/repository"
	"test/internal/services"
)

func main() {

	tasksChan := make(chan models.Task, 5)

	var wg sync.WaitGroup
	var taskBuffer []*models.Task

	cfg := app.NewConfig()
	mutex := &sync.Mutex{}

	app := app.NewApp(&wg, cfg, taskBuffer, &tasksChan, mutex)

	//Подключаем воркеров
	for i := 1; i <= cfg.WorkersNum; i++ {
		app.WG.Add(1)
		go services.Worker(i, *app.TaskChannel, app)
	}

	//Подгружаем сохраненные раннее задачи
	LoadTasks(app)

	mux := http.NewServeMux()

	linkRepo := repository.NewLinkRepostiory(app)
	linkService := services.NewLinkService(linkRepo, app)

	taskRepo := repository.NewTaskRepository(app)
	taskService := services.NewTaskService(taskRepo, app)
	handler := httpHandlers.NewLinkHandler(app, taskService, linkService)

	mux.HandleFunc("POST /links", handler.ProcessLinkHandler)
	mux.HandleFunc("GET /shutdown", handler.ShutDown)

	fmt.Println("Server starting on port 8080...")

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func LoadTasks(a *app.App) {

	path := filepath.Join(a.Config.DataDir, "tasksQueue.json")

	data, err := repository.ReadJson[models.Task](path)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла:%v\n", err)
	}

	for _, el := range data {
		select {
		case *a.TaskChannel <- el:
			fmt.Printf("Задача %s была загружена из памяти в очередь\n", el.ID)
		default:
			a.TaskBuffer = append(a.TaskBuffer, &el)
			fmt.Printf("Задача %s была загружена из памяти в буфер\n", el.ID)
		}
	}

	//Очищаем файл с сохраненными задачами
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Ошибка открытия файла:", err)
		return
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		fmt.Println("Ошибка обрезки файла:", err)
		return
	}
}
