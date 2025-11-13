package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"test/internal/models"
)

type App struct {
	Draining    atomic.Bool
	WG          *sync.WaitGroup
	Config      *Config
	TaskBuffer  []*models.Task
	TaskChannel *chan models.Task
	Mutex       *sync.Mutex
}

func NewApp(wg *sync.WaitGroup, config *Config, taskBuffer []*models.Task, taskChannel *chan models.Task, mu *sync.Mutex) *App {
	return &App{
		Draining:    atomic.Bool{},
		WG:          wg,
		Config:      config,
		TaskBuffer:  taskBuffer,
		TaskChannel: taskChannel,
		Mutex:       mu,
	}
}

func (a *App) FlushBuffer() {
	a.Mutex.Lock()
	defer a.Mutex.Unlock()

	count := 0
	for len(a.TaskBuffer) > 0 {
		select {
		case *a.TaskChannel <- *a.TaskBuffer[0]:
			fmt.Printf("Слили задачу %s\n", a.TaskBuffer[0].ID)
			a.TaskBuffer = a.TaskBuffer[1:] // удаляем первую задачу
			count++
		default:
			fmt.Println("В буфере сейчас:")
			for _, task := range a.TaskBuffer {
				fmt.Printf("%s\n", task.ID)
			}
			return
		}
	}

}
