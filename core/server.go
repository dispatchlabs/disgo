package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"github.com/dispatchlabs/disgo/properties"
	"github.com/dispatchlabs/disgo_commons/types"
	log "github.com/sirupsen/logrus"
	dapos "github.com/dispatchlabs/dapos/core"
	disgover "github.com/dispatchlabs/disgover/core"
	"github.com/dispatchlabs/disgo_commons/services"
)

const (
	Version = "1.0.0"
)

// Server
type Server struct {
	services   []types.IService
	api        *Api
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

// Go
func (server *Server) Go() {
	log.Info("booting Disgo v" + Version + "...")
	log.Info("args  [" + strings.Join(os.Args, " ") + "]")

	/*
	transaction := types.NewTransaction()
	transaction.From, _ = crypto.NewWalletAddress()
	log.Info(transaction.String())
	*/

	// Add services.
	server.services = append(server.services, dapos.NewDAPoSService())
	server.services = append(server.services, disgover.NewDisGoverService())
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

