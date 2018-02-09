package server

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/dispatchlabs/disgo/properties"
	"github.com/dispatchlabs/disgo/services"
	"github.com/dispatchlabs/disgover"
	log "github.com/sirupsen/logrus"
)

type Server struct {
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

	seedNodes := []*disgover.Contact{}

	if len(os.Args) > 1 && os.Args[1] == "-seed" {
		seedNodes = append(seedNodes,
			&disgover.Contact{
				Id: "NODE-Seed-001",
				Endpoint: &disgover.Endpoint{
					Port: int64(properties.Properties.GrpcPort),
				},
			},
		)
	}

	disgover.SetInstance(
		disgover.NewDisgover(
			disgover.NewContact(),
			seedNodes,
		),
	)
	disgover.GetInstance().ThisContact.Endpoint.Port = int64(properties.Properties.GrpcPort)

	if (len(os.Args) > 1) && (strings.Index(os.Args[1], "-nodeId=") == 0) {
		var nodeID = strings.Replace(os.Args[1], "-nodeId=", "", -1)
		disgover.GetInstance().ThisContact.Id = nodeID
	}

	// Run
	var waitGroup sync.WaitGroup
	server.services = append(server.services, services.NewHttpService())
	server.services = append(server.services, services.NewGrpcService())
	server.services = append(server.services, services.NewDisgoverService())
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
