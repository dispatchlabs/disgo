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
	NodeId   string
	ThisIp   string
	SeedList []string
}

func (server *Server) Start() {
	log.Info("Args: " + strings.Join(os.Args, " "))

	var cmdParams = CmdParams{}

	// Parse CMD Args
	for _, arg := range os.Args {
		if strings.Index(arg, "-nodeId=") == 0 {
			cmdParams.NodeId = strings.Replace(arg, "-nodeId=", "", -1)
		} else if strings.Index(arg, "-thisIp=") == 0 {
			cmdParams.ThisIp = strings.Replace(arg, "-thisIp=", "", -1)
		} else if strings.Index(arg, "-seedList=") == 0 {
			var seedList = strings.Replace(arg, "-seedList=", "", -1)
			cmdParams.SeedList = strings.Split(seedList, ";")
		}
	}

	// Set THIS Contact/Node on the network
	var thisContact = disgover.NewContact()
	if len(cmdParams.NodeId) > 0 {
		thisContact.Id = cmdParams.NodeId
	}
	if len(cmdParams.ThisIp) > 0 {
		thisContact.Endpoint.Host = cmdParams.ThisIp
	}
	thisContact.Endpoint.Port = int64(properties.Properties.GrpcPort)

	// Check if we have a seed list
	var seedList = []*disgover.Contact{}
	for _, seedIP := range cmdParams.SeedList {
		seedList = append(seedList, &disgover.Contact{
			Endpoint: &disgover.Endpoint{
				Host: seedIP,
				Port: int64(properties.Properties.GrpcPort),
			},
		})
	}

	// Instantiate the node
	disgover.SetInstance(disgover.NewDisgover(thisContact, seedList))

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
