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

	"github.com/google/uuid"
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
	ctx := r.Context()

	var data dto.LinkListRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		fmt.Println("неверный формат", r.Body)
		return
	}

	//Парсим данные в нужную структуру
	linkID := uuid.New().String()

	linkMap := make(map[string]string)

	for _, el := range data.Links {
		linkMap[el] = "not available"
	}

	resultChan := make(chan interface{}, 1)

	task := &domain.TaskDomain{
		ID:         uuid.New().String(), //ID задачи
		LinkID:     []string{linkID},    //ID набора линков
		LinksSets:  linkMap,             //Мапа с URL:STATUS //{"google.com":"not available"}
		TaskType:   domain.CheckURL,     //Тип таски: проверка URL или получение PDF
		Ctx:        ctx,                 //Контекст
		ResultChan: resultChan,          //Канал для результата
	}

	//Создаем и запускаем в канал таску, ожидаем ответ
	result, err := lh.TaskService.CreateTaskForLinkService(ctx, task)
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
	ctx := r.Context()

	var data dto.PdfRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	//Инициализируем канал и задачу
	resultChan := make(chan interface{}, 1)

	task := &domain.TaskDomain{
		ID:         uuid.New().String(),
		LinkID:     data.LinksIDS,
		LinksSets:  nil,
		TaskType:   domain.LoadPDF,
		Ctx:        ctx,
		ResultChan: resultChan,
	}

	//Создаем и запускаем в канал таску, ожидаем ответ
	result, err := lh.TaskService.CreateTaskForLinkService(ctx, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if result == "" {
		http.Error(w, "не удалось создать файл", http.StatusInternalServerError)
		return
	}
	//Формируем заголовки и читаем PDF для передачи
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

	//Пауза для отладки
	//time.Sleep(10 * time.Second)

	//Закрываем канал с тасками
	close(*lh.App.TaskChannel)

	// ждём завершения воркеров
	lh.App.WG.Wait()

	fmt.Fprint(w, "Сервер завершил работу")
}
