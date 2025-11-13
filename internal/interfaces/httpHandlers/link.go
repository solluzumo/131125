package httpHandlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"test/internal/app"
	"test/internal/dto"
	"test/internal/models"
	"test/internal/services"
)

func generateShortID(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type LinkHandler struct {
	IsDraining     func() bool
	UpdateDraining func()
	App            *app.App
	TaskService    *services.TaskService
}

func NewLinkHandler(app *app.App, tService *services.TaskService) *LinkHandler {
	return &LinkHandler{
		IsDraining:     func() bool { return app.Draining.Load() },
		UpdateDraining: func() { app.Draining.Store(true) },
		TaskService:    tService,
		App:            app,
	}
}

func (lh *LinkHandler) ProcessLinkHandler(w http.ResponseWriter, r *http.Request) {

	var data dto.LinkListRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	task := &models.Task{
		ID:   generateShortID(6),
		Data: data.Links,
	}

	if lh.IsDraining() {
		taskArray := make([]*models.Task, 0)
		taskArray = append(taskArray, task)

		fmt.Fprint(w, "Сервер начал остановку, ваша задача поставлена в очередь")

		if !lh.TaskService.SaveTaskService(taskArray) {
			http.Error(w, "Не удалось сохранить вашу задачу", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, "\nЗадача сохранена")
		return
	}
	select {
	case lh.App.TaskChannel <- *task:
		fmt.Printf("Задача %s в обработке!\n", task.ID)
	default:
		fmt.Printf("Задача %s добавлена в буфер!\n", task.ID)
		lh.App.TaskBuffer = append(lh.App.TaskBuffer, task)
	}

}

func (lh *LinkHandler) ShutDown(w http.ResponseWriter, r *http.Request) {
	if lh.IsDraining() {
		http.Error(w, "Сервер уже остановлен!", http.StatusBadRequest)
		return
	}

	fmt.Println("Сервер начал остановку!")

	lh.UpdateDraining()       // флаг "draining"
	close(lh.App.TaskChannel) // закрыли канал, новых задач не будет

	var leftover []*models.Task

	// читаем всё, что осталось в канале
	for task := range lh.App.TaskChannel {
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
