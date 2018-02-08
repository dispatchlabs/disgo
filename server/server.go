package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/dispatchlabs/disgo/properties"
	"github.com/dispatchlabs/disgo/services"
	"github.com/dispatchlabs/disgover"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Disgover *disgover.Disgover
	services []services.IService
}

func NewServer() *Server {

	// Setup log.
	formatter := &log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   false,
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

	// Kubernetes - Intervention Start
	var seedNodeIP = os.Getenv("SEED_NODE_IP") // Needed when run from Kubernetes
	var rootNodeId = "NODE-1"

	seedNodes := []*disgover.Contact{}
	if len(seedNodeIP) != 0 {
		seedNodes = append(seedNodes,
			&disgover.Contact{
				Id: rootNodeId,
				Endpoint: &disgover.Endpoint{
					Host: seedNodeIP,
					Port: 1975,
				},
			},
		)
	}
	server.Disgover = disgover.NewDisgover(
		disgover.NewContact(),
		seedNodes,
	)
	disgover.DisgoverSingleton = server.Disgover

	if len(seedNodeIP) == 0 {
		server.Disgover.ThisContact.Id = rootNodeId
	}
	server.Disgover.ThisContact.Endpoint.Port = int64(properties.Properties.GrpcPort)
	// Kubernetes - Intervention End

	// Run
	var waitGroup sync.WaitGroup
	server.services = append(server.services, services.NewHttpService())
	server.services = append(server.services, services.NewGrpcService(server.Disgover))
	server.services = append(server.services, services.NewDisgoverService(server.Disgover))
	server.services = append(server.services, services.NewHelloService())

	for _, service := range server.services {
		log.Info("starting " + service.Name() + "...")
		go service.Go(&waitGroup)
	}

	for i := 0; i < len(server.services); i++ {
		waitGroup.Add(1)
	}

	waitGroup.Wait()
}
