package app

import (
	"sync"
	"sync/atomic"
	"test/internal/domain"
)

type App struct {
	Draining    atomic.Bool
	WG          *sync.WaitGroup
	Config      *Config
	TaskChannel *chan domain.TaskDomain
	Mutex       *sync.Mutex
}

func NewApp(wg *sync.WaitGroup, config *Config, taskBuffer []*domain.TaskDomain, taskChannel *chan domain.TaskDomain, mu *sync.Mutex) *App {
	return &App{
		Draining:    atomic.Bool{},
		WG:          wg,
		Config:      config,
		TaskChannel: taskChannel,
		Mutex:       mu,
	}
}
