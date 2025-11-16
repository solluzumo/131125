package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	mux.HandleFunc("POST /links", handler.ProcessLinkHandler)
	mux.HandleFunc("POST /get-pdf", handler.ProcessPDFHandler)
	mux.HandleFunc("GET /shutdown", handler.ShutDown)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		fmt.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ждём Ctrl+C
	<-ctx.Done()
	fmt.Println("\nShutting down...")

	// Завершаем HTTP-сервер
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	close(tasksChan)
	wg.Wait()

	fmt.Println("Server exited properly")

}
