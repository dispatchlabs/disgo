package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"github.com/dispatchlabs/disgo/properties"
	"github.com/dispatchlabs/disgo_commons/types"
	"github.com/dispatchlabs/disgover"
	log "github.com/sirupsen/logrus"
	"github.com/dispatchlabs/disgo_commons/services"
	"github.com/dispatchlabs/dapos"
)

const (
	Version = "1.0.0"
)

// Server
type Server struct {
	services []types.IService
	api      *Api
}

// NewServer
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
	fileName := "." + string(os.PathSeparator) + "properties" + string(os.PathSeparator) + "disgo.json"
	file, error := ioutil.ReadFile(fileName)
	if error != nil {
		log.Error("unable to load " + fileName + "[error=" + error.Error() + "]")
		os.Exit(1)
	}
	json.Unmarshal(file, &properties.Properties)

	// Load Keys
	if _, _, err := loadKeys(); err != nil {
		log.Error("unable to keys: " + err.Error())
	}

	return &Server{}
}

// Go
func (server *Server) Go() {
	log.Info("booting Disgo v" + Version + "...")

	// Add services.
	server.services = append(server.services, dapos.NewDAPoSService().WithGrpc())
	server.services = append(server.services, disgover.NewDisGoverService().WithGrpc())
	server.services = append(server.services, services.NewStoreService())
	server.services = append(server.services, services.NewHttpService())
	server.services = append(server.services, services.NewGrpcService())

	// Create api.
	server.api = NewApi(server.services)

	// Run services.
	var waitGroup sync.WaitGroup
	for _, service := range server.services {
		log.WithFields(log.Fields{
			"method": "Server.Go",
		}).Info("starting " + service.Name() + "...")
		go service.Go(&waitGroup)
		waitGroup.Add(1)
	}
	waitGroup.Wait()
}
