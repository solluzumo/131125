package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"test/internal/app"
	"test/internal/domain"
	"test/internal/interfaces/httpHandlers"
	"test/internal/repository"
	"test/internal/services"
	"time"
)

func main() {

	cfg := app.NewConfig()

	mutex := &sync.Mutex{}
	tasksChan := make(chan domain.TaskDomain, cfg.TaskChanCap)

	var wg sync.WaitGroup
	var taskBuffer []*domain.TaskDomain

	app := app.NewApp(&wg, cfg, taskBuffer, &tasksChan, mutex)

	mux := http.NewServeMux()

	linkRepo := repository.NewLinkRepostiory(app, app.Mutex, cfg.JsonDir, cfg.PdfDir)
	linkService := services.NewLinkService(linkRepo)

	taskRepo := repository.NewTaskRepository(app, cfg.JsonDir, app.Mutex)
	taskService := services.NewTaskService(taskRepo, linkRepo, app)
	handler := httpHandlers.NewLinkHandler(app, taskService, linkService)

	//Подключаем воркеров
	for i := 1; i <= cfg.WorkersNum; i++ {
		app.WG.Add(1)
		go services.Worker(i, *app.TaskChannel, taskService, app.WG)
	}

	//Подгружаем сохраненные раннее задачи
	LoadTasks(app, taskRepo)

	mux.HandleFunc("POST /links", handler.ProcessLinkHandler)
	mux.HandleFunc("POST /get-pdf", handler.ProcessPDFHandler)
	mux.HandleFunc("GET /shutdown", handler.ShutDown)

	fmt.Println("Server starting on port 8080...")

	app.StartFlushBufferTicker(2 * time.Second)

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func LoadTasks(a *app.App, taskRepo *repository.TaskRepostiory) {

	data, err := taskRepo.ReadTaskJson()
	if err != nil {
		fmt.Printf("Ошибка при чтении файла:%v\n", err)
	}

	for _, el := range data {
		if el.TaskStatus == string(domain.Done) {
			continue
		}
		resultChan := make(chan interface{}, 1)
		taskObj := domain.TaskDomain{
			ID:         el.ID,
			LinksID:    el.LinksID,
			TaskType:   domain.TaskType(el.TaskType),
			TaskStatus: domain.TaskStatus(el.TaskStatus),
			ResultChan: resultChan,
		}
		select {
		case *a.TaskChannel <- taskObj:
			fmt.Printf("Задача %s была загружена из памяти в очередь\n", el.ID)
		default:
			a.TaskBuffer = append(a.TaskBuffer, &taskObj)
			fmt.Printf("Задача %s была загружена из памяти в буфер\n", el.ID)
		}

	}

}
