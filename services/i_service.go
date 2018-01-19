package services

import "sync"

type IService interface {
	Name() string
	IsRunning() bool
	Go(waitGroup *sync.WaitGroup)
}
