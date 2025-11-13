package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"test/internal/models"
)

type Config struct {
	DataDir    string
	WorkersNum int
}

type App struct {
	Draining    atomic.Bool
	WG          *sync.WaitGroup
	Config      *Config
	TaskBuffer  []*models.Task
	TaskChannel chan models.Task
	mu          sync.Mutex
}

func NewConfig() *Config {
	return &Config{
		DataDir:    "./data",
		WorkersNum: 2,
	}
}

func (a *App) FlushBuffer() {
	a.mu.Lock()
	defer a.mu.Unlock()

	count := 0
	for len(a.TaskBuffer) > 0 {
		select {
		case a.TaskChannel <- *a.TaskBuffer[0]:
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
