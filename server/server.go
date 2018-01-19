package server

import (
	"github.com/dispatchlabs/disgo_node/services"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"github.com/dispatchlabs/disgo_node/configurations"
)

type Server struct {
	services [] services.IService
}

func NewServer() *Server {

	server := &Server{}
	filePath := "." + string(os.PathSeparator) + "configurations" + string(os.PathSeparator) + "disgo_node_config.json"

	// Read configuration JSON file.
	file, error := ioutil.ReadFile(filePath)
	if error != nil {
		log.Error("unable to load " + filePath)
		os.Exit(1)
	}
	json.Unmarshal(file, &configurations.Configuration)

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
