package httpHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"test/internal/app"
	"test/internal/dto"
	"test/internal/models"
	"test/internal/services"
)

type LinkHandler struct {
	IsDraining     func() bool
	UpdateDraining func()
	App            *app.App
	TaskService    *services.TaskService
	LinkService    *services.LinkService
}

func NewLinkHandler(app *app.App, tService *services.TaskService, lService *services.LinkService) *LinkHandler {
	return &LinkHandler{
		IsDraining:     func() bool { return app.Draining.Load() },
		UpdateDraining: func() { app.Draining.Store(true) },
		TaskService:    tService,
		App:            app,
		LinkService:    lService,
	}
}

func (lh *LinkHandler) ProcessLinkHandler(w http.ResponseWriter, r *http.Request) {

	var data dto.LinkListRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	linkList := lh.LinkService.SaveLinkService(data)

	result, err := lh.TaskService.CreateTaskForLinkService(linkList.ID, models.CheckURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (lh *LinkHandler) ShutDown(w http.ResponseWriter, r *http.Request) {
	if lh.IsDraining() {
		http.Error(w, "Сервер уже остановлен!", http.StatusBadRequest)
		return
	}

	fmt.Println("Сервер начал остановку!")

	lh.UpdateDraining()        // флаг "draining"
	close(*lh.App.TaskChannel) // закрыли канал, новых задач не будет

	var leftover []*models.Task

	// читаем всё, что осталось в канале
	for task := range *lh.App.TaskChannel {
		fmt.Printf("Сохраняем %s\n", task.ID)
		leftover = append(leftover, &task)
	}

	//Сохраняем таски из буфера
	leftover = append(leftover, lh.App.TaskBuffer...)

	// ждём завершения воркеров
	lh.App.WG.Wait()

	// сохраняем невыполненные задачи
	lh.TaskService.SaveTaskService(leftover)

	fmt.Fprint(w, "Сервер завершил работу")
}
