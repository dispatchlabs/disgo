package services

import (
	"net"
	"strconv"
	"sync"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/dispatchlabs/disgo/properties"
	"github.com/dispatchlabs/disgo/party"
	"github.com/dispatchlabs/disgover"
)

type GrpcService struct {
	Port     int
	Disgover *disgover.Disgover
	running  bool
}

func NewGrpcService(disgover *disgover.Disgover) *GrpcService {
	return &GrpcService{properties.Properties.GrpcPort, disgover, false}
}

func (grpcService *GrpcService) Name() string {
	return "GrpcService"
}

func (grpcService *GrpcService) IsRunning() bool {
	return grpcService.running
}

func (grpcService *GrpcService) Go(waitGroup *sync.WaitGroup) {

	grpcService.running = true
	listen, error := net.Listen("tcp", ":"+strconv.Itoa(grpcService.Port))
	if error != nil {
		log.Fatalf("failed to listen: %v", error)
	}
	grpcServer := grpc.NewServer()

	// Register disgoGrpc.
	log.WithFields(log.Fields{
		"method": grpcService.Name() + ".Go",
	}).Info("registering Disgover...")
	party.RegisterPartyServer(grpcServer, party.NewParty())

	// Register Disgover.
	log.WithFields(log.Fields{
		"method": grpcService.Name() + ".Go",
	}).Info("registering Disgover...")
	disgover.RegisterDisgoverRPCServer(grpcServer, grpcService.Disgover)

	// Serve.
	reflection.Register(grpcServer)
	log.WithFields(log.Fields{
		"method": grpcService.Name() + ".Go",
	}).Info("listening on " + strconv.Itoa(grpcService.Port))
	if error := grpcServer.Serve(listen); error != nil {
		log.Fatalf("failed to serve: %v", error)
		grpcService.running = false
	}
}
