package httpHandlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"test/internal/app"
	"test/internal/domain"
	"test/internal/dto"
	"test/internal/services"
	"time"
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

// Функция обработки запроса на проверку линков
func (lh *LinkHandler) ProcessLinkHandler(w http.ResponseWriter, r *http.Request) {

	var data dto.LinkListRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		fmt.Println("неверный формат", r.Body)
		return
	}
	//Сохраняем линки
	linkList := lh.LinkService.SaveLinkService(data)

	linkIdSlice := []string{linkList.ID}

	//Создаем и запускаем в канал таску, ожидаем ответ
	result, err := lh.TaskService.CreateTaskForLinkService(linkIdSlice, domain.CheckURL)
	if err != nil {
		//Если сервис в процессе отключения
		if err.Error() == "draining" {
			response := fmt.Sprintf("Сервер в процессе отключения, обратитесь по адресу localhost://get-result/%s", result)
			fmt.Fprint(w, response)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Функция обработки запроса на создания PDF из наборов линков
func (lh *LinkHandler) ProcessPDFHandler(w http.ResponseWriter, r *http.Request) {

	var data dto.PdfRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	//Создаем и запускаем в канал таску, ожидаем ответ
	result, err := lh.TaskService.CreateTaskForLinkService(data.Links, domain.LoadPDF)
	if err != nil {
		if err.Error() == "draining" {
			response := fmt.Sprintf("Сервер в процессе отключения, обратитесь по адресу localhost://get-result/%s", result)
			fmt.Fprint(w, response)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	file, err := os.Open(result.(string))
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"simple.pdf\"")

	_, _ = io.Copy(w, file)

}

// Функция обработки запроса на отключение сервера
func (lh *LinkHandler) ShutDown(w http.ResponseWriter, r *http.Request) {
	//Если сервер уже остановлен
	if lh.IsDraining() {
		http.Error(w, "Сервер уже остановлен!", http.StatusBadRequest)
		return
	}

	fmt.Println("Сервер начал остановку!")

	//Меняем статус сервера на отключение
	lh.UpdateDraining()

	//Пауза для отладки, можно убрать
	time.Sleep(10 * time.Second)

	//Закрываем канал с тасками
	close(*lh.App.TaskChannel)

	var leftover []*domain.TaskDomain

	// читаем всё, что осталось в канале
	for task := range *lh.App.TaskChannel {
		fmt.Printf("Сохраняем %s\n", task.ID)
		task.TaskStatus = domain.FromMemory
		leftover = append(leftover, &task)
	}

	//Сохраняем таски из буфера
	for _, task := range lh.App.TaskBuffer {
		fmt.Printf("Сохраняем %s\n", task.ID)
		task.TaskStatus = domain.FromMemory
		leftover = append(leftover, task)
	}

	// ждём завершения воркеров
	lh.App.WG.Wait()

	// сохраняем невыполненные задачи
	lh.TaskService.SaveTaskService(leftover)

	fmt.Fprint(w, "Сервер завершил работу")
}
