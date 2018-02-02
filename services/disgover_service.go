package services

import (
	"sync"
	"github.com/dispatchlabs/disgover"
)

type DisgoverService struct {
	Disgover *disgover.Disgover
	running bool
}

func NewDisgoverService(disgover *disgover.Disgover) *DisgoverService {
	disgoverService := DisgoverService{disgover,false}
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
	disgoverService.Disgover.Go()
}
