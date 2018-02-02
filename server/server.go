package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"github.com/dispatchlabs/disgo/services"
	log "github.com/sirupsen/logrus"
	"github.com/dispatchlabs/disgo/properties"
	"github.com/dispatchlabs/disgover"
)


type Server struct {
	Disgover *disgover.Disgover
	services []services.IService
}

func NewServer() *Server {

	// Setup log.
	formatter := &log.TextFormatter{
		FullTimestamp: true,
		ForceColors: false,
	}
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Read configuration JSON file.
	filePath := "." + string(os.PathSeparator) + "properties" + string(os.PathSeparator) + "disgo.json"
	file, error := ioutil.ReadFile(filePath)
	if error != nil {
		log.Error("unable to load " + filePath)
		os.Exit(1)
	}
	json.Unmarshal(file, &properties.Properties)

	// Load Keys
	if _, _, err := loadKeys(); err != nil {
		log.Error("unable to keys: " + err.Error())
	}

	return &Server{}
}

func (server *Server) Start() {

	// Create Disgover.
	server.Disgover = disgover.NewDisgover(
		disgover.NewContact(),
		[]*disgover.Contact{},
	)

	var waitGroup sync.WaitGroup
	server.services = append(server.services, services.NewHttpService())
	server.services = append(server.services, services.NewGrpcService(server.Disgover))
	server.services = append(server.services, services.NewDisgoverService(server.Disgover))

	for _, service := range server.services {
		log.Info("starting " + service.Name() + "...")
		go service.Go(&waitGroup)
	}

	for i := 0; i < len(server.services); i++ {
		waitGroup.Add(1)
	}

	waitGroup.Wait()
}
