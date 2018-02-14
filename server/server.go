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

type CmdParams struct {
	IsSeed   bool
	NodeName string
}

func (server *Server) Start() {
	log.Info("Args: " + strings.Join(os.Args, " "))

	var cmdParams = CmdParams{}

	for _, arg := range os.Args {
		if arg == "-seed" {
			cmdParams.IsSeed = true
		}
		if strings.Index(arg, "-nodeId=") == 0 {
			cmdParams.NodeName = strings.Replace(arg, "-nodeId=", "", -1)
		}
	}

	var seedContact = &disgover.Contact{
		Id: cmdParams.NodeName,
		Endpoint: &disgover.Endpoint{
			Host: "35.227.162.40",
			Port: int64(properties.Properties.GrpcPort),
		},
	}
	var thisContact = disgover.NewContact()
	thisContact.Endpoint.Port = int64(properties.Properties.GrpcPort)

	if len(cmdParams.NodeName) > 0 {
		thisContact.Id = cmdParams.NodeName
	}

	if cmdParams.IsSeed {
		disgover.SetInstance(
			disgover.NewDisgover(
				seedContact,
				[]*disgover.Contact{},
			),
		)
	} else {
		disgover.SetInstance(
			disgover.NewDisgover(
				thisContact,
				[]*disgover.Contact{
					seedContact,
				},
			),
		)

	}

	// Run
	var waitGroup sync.WaitGroup
	// server.services = append(server.services, services.NewHttpService())
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
