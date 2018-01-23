package server

import (
	"github.com/dispatchlabs/disgo/services"
	log "github.com/sirupsen/logrus"
	"sync"
	"os"
	"io/ioutil"
	"encoding/json"
	"github.com/dispatchlabs/disgo/configs"
)

type Server struct {
	services [] services.IService
}

func NewServer() *Server {

	// Setup log.
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Read configuration JSON file.
	filePath := "." + string(os.PathSeparator) + "configs" + string(os.PathSeparator) + "disgo_config.json"
	file, error := ioutil.ReadFile(filePath)
	if error != nil {
		log.Error("unable to load " + filePath)
		os.Exit(1)
	}
	json.Unmarshal(file, &configs.Config)

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
