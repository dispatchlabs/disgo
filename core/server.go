package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/dispatchlabs/dapos"
	"github.com/dispatchlabs/disgo_commons/config"
	"github.com/dispatchlabs/disgo_commons/services"
	"github.com/dispatchlabs/disgo_commons/types"
	"github.com/dispatchlabs/disgo_commons/utils"
	"github.com/dispatchlabs/disgover"
	log "github.com/sirupsen/logrus"
)

const (
	Version = "1.0.0"
)

// Server
type Server struct {
	services []types.IService
	api      *Api
}

// NewServer -
func NewServer() *Server {

	// Setup log.
	formatter := &log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   false,
	}
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// Read configuration JSON file and override default values
	config.Properties = &config.DisgoProperties{
		HttpPort:          1975,
		HttpHostIp:        "0.0.0.0",
		GrpcPort:          1973,
		GrpcTimeout:       5,
		UseQuantumEntropy: false,
		IsSeed:            false,
		IsDelegate:        false,
		SeedList:          []string{},
		DaposDelegates:    []string{},
		NodeId:            "",
		ThisIp:            "",
	}

	var configFile = utils.GetDisgoDir() + string(os.PathSeparator) + "config.json"
	if utils.Exists(configFile) {
		file, error := ioutil.ReadFile(configFile)
		if error != nil {
			log.Error("unable to load " + configFile + "[error=" + error.Error() + "]")
			os.Exit(1)
		}
		json.Unmarshal(file, &config.Properties)
	}

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
	if !config.Properties.IsSeed {
		server.services = append(server.services, NewPingPongService())
	}
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
