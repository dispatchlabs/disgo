package server

import (
	"github.com/dispatchlabs/disgo/services"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"github.com/dispatchlabs/disgo/configurations"
)

type Server struct {
	services [] services.IService
}

func NewServer() *Server {

	server := &Server{}

	return server
}

func (server *Server) Start() {

	var waitGroup sync.WaitGroup
	server.services = append(server.services, services.NewHttpService())
	server.services = append(server.services, services.NewRpcService())

	for _, service := range server.services {
		log.Info("starting " + service.Name() + "...")
		go service.Go(&waitGroup)
	}

	for i := 0; i < len(server.services); i++ {
		waitGroup.Add(1)
	}

	waitGroup.Wait()
}
