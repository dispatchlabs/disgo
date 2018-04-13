package core

import (
	"sync"

	"github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/dapos"
	"github.com/dispatchlabs/disgover"
)

const (
	Version = "1.0.0"
)

// Server -
type Server struct {
	services []types.IService
}

// NewServer -
func NewServer() *Server {
	utils.InitializeLogger()

	// Load Keys
	if _, _, err := loadKeys(); err != nil {
		utils.Error("unable to keys: " + err.Error())
	}

	return &Server{}
}

// Go
func (server *Server) Go() {
	utils.Info("booting Disgo v" + Version + "...")

	// Add services.
	// if !config.Properties.IsSeed {
	// 	server.services = append(server.services, NewPingPongService())
	// }
	server.services = append(server.services, services.GetDbService())
	server.services = append(server.services, disgover.GetDisGoverService().WithGrpc().WithHttp())
	server.services = append(server.services, dapos.GetDAPoSService().WithGrpc().WithHttp())
	server.services = append(server.services, services.GetHttpService())
	server.services = append(server.services, services.GetGrpcService())

	// Run services.
	var waitGroup sync.WaitGroup
	for _, service := range server.services {
		utils.Info("starting " + utils.GetStructName(service) + "...")
		go service.Go(&waitGroup)
		waitGroup.Add(1)
	}
	waitGroup.Wait()
}
