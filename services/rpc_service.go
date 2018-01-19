package services

import (
	"sync"
	"github.com/dispatchlabs/disgo_node/configurations"
)

type RpcService struct {
	Port int
	running bool
}

func NewRpcService() *RpcService {

	rpcService := RpcService{configurations.Configuration.RpcPort, false}

	return &rpcService
}

func (rpcService *RpcService) Name() string {
	return "RpcService"
}

func (rpcService *RpcService) IsRunning() bool {
	return rpcService.running
}

func (rpcService *RpcService) Go(waitGroup *sync.WaitGroup) {
	rpcService.running = true
}
