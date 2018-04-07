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
		file, err := ioutil.ReadFile(configFileName)
		if err != nil {
			log.Error("unable to load " + configFileName + "[error=" + err.Error() + "]")
			os.Exit(1)
		}
		json.Unmarshal(file, &config.Properties)
	} else {
		file, err := os.Create(configFileName)
		defer file.Close()
		if err != nil {
			utils.Error("unable to create " + configFileName + " [error=" + err.Error() + "]")
			panic(err)
		}
		bytes, err := json.Marshal(&config.Properties)
		if err != nil {
			utils.Error("unable to Marshal Properties [error=" + err.Error() + "]")
			panic(err)
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
	utils.Info("booting Disgo v" + Version + "...")

	// Add services.
	// if !config.Properties.IsSeed {
	// 	server.services = append(server.services, NewPingPongService())
	// }
	server.services = append(server.services, services.GetDbService())
	server.services = append(server.services, disgover.GetDisGoverService().WithGrpc().WithHttp())
	server.services = append(server.services, dapos.GetDAPoSService().WithGrpc().WithHttp())
	server.services = append(server.services, services.GetHttpService())
	server.services = append(server.services, services.GetGrpcService())

	// Run services.
	var waitGroup sync.WaitGroup
	for _, service := range server.services {
		utils.Info("starting " + utils.GetStructName(service) + "...")
		go service.Go(&waitGroup)
		waitGroup.Add(1)
	}
	waitGroup.Wait()
}
