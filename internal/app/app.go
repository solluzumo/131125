package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"test/internal/domain"
	"time"
)

type App struct {
	Draining    atomic.Bool
	WG          *sync.WaitGroup
	Config      *Config
	TaskBuffer  []*domain.TaskDomain
	TaskChannel *chan domain.TaskDomain
	Mutex       *sync.Mutex
}

func NewApp(wg *sync.WaitGroup, config *Config, taskBuffer []*domain.TaskDomain, taskChannel *chan domain.TaskDomain, mu *sync.Mutex) *App {
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
			a.TaskBuffer = a.TaskBuffer[1:]
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

func (a *App) StartFlushBufferTicker(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			a.FlushBuffer()
		}
	}()
}
