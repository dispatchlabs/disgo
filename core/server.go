package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/dispatchlabs/commons/config"
	"github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/commons/utils"
	"github.com/dispatchlabs/dapos"
	"github.com/dispatchlabs/disgover"
	log "github.com/sirupsen/logrus"
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
		SeedList: []string{
			"35.230.30.125",
		},
		DaposDelegates: []string{
			"test-net-delegate-1",
			"test-net-delegate-2",
			"test-net-delegate-3",
			"test-net-delegate-4",
		},
		NodeId: "",
		ThisIp: "",
	}

	log.WithFields(log.Fields{
		"method": utils.GetCallingFuncName() + fmt.Sprintf(" -> %s", utils.GetDisgoDir()),
	}).Info("config folder")

	var configFileName = utils.GetDisgoDir() + string(os.PathSeparator) + "config.json"
	if utils.Exists(configFileName) {
		file, error := ioutil.ReadFile(configFileName)
		if error != nil {
			log.Error("unable to load " + configFileName + "[error=" + error.Error() + "]")
			os.Exit(1)
		}
		json.Unmarshal(file, &config.Properties)
	} else {
		file, error := os.Create(configFileName)
		defer file.Close()
		if error != nil {
			log.WithFields(log.Fields{
				"method": utils.GetCallingFuncName(),
			}).Fatal("unable to create " + configFileName + " [error=" + error.Error() + "]")
			panic(error)
		}
		bytes, error := json.Marshal(&config.Properties)
		if error != nil {
			log.WithFields(log.Fields{
				"method": utils.GetCallingFuncName(),
			}).Fatal("unable to Marshal Properties [error=" + error.Error() + "]")
			panic(error)
		}
		fmt.Fprintf(file, string(bytes))
	}

	// Load Keys
	if _, _, err := loadKeys(); err != nil {
		log.Error("unable to keys: " + err.Error())
	}

	return &Server{}
}

// Go
func (server *Server) Go() {
	log.WithFields(log.Fields{
		"method": utils.GetCallingFuncName(),
	}).Info("booting Disgo v" + Version + "...")

	// Add services.
	// if !config.Properties.IsSeed {
	// 	server.services = append(server.services, NewPingPongService())
	// }
	server.services = append(server.services, disgover.NewDisGoverService().WithGrpc())
	server.services = append(server.services, dapos.NewDAPoSService().WithGrpc())
	server.services = append(server.services, services.NewHttpService())
	server.services = append(server.services, services.NewGrpcService())

	// Register handlers.
	registerHttpHandlers()

	// Run services.
	var waitGroup sync.WaitGroup
	for _, service := range server.services {
		log.WithFields(log.Fields{
			"method": utils.GetCallingFuncName(),
		}).Info("starting " + utils.GetStructName(service) + "...")
		go service.Go(&waitGroup)
		waitGroup.Add(1)
	}
	waitGroup.Wait()
}
