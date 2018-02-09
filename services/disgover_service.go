package services

import (
	"sync"

	"github.com/dispatchlabs/disgover"
)

type DisgoverService struct {
	running bool
}

func NewDisgoverService() *DisgoverService {
	disgoverService := DisgoverService{false}
	return &disgoverService
}

func (disgoverService *DisgoverService) Name() string {
	return "DisgoverService"
}

func (disgoverService *DisgoverService) IsRunning() bool {
	return disgoverService.running
}

func (disgoverService *DisgoverService) Go(waitGroup *sync.WaitGroup) {
	disgoverService.running = true
	disgover.GetInstance().Go()
}
