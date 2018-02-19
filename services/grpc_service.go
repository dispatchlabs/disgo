package services

import (
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"sync"
	"github.com/dispatchlabs/disgo/properties"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcService
type GrpcService struct {
	running    bool
	grpcServer *grpc.Server
}

// NewGrpcService
func NewGrpcService(grpcServer *grpc.Server) *GrpcService {
	return &GrpcService{
		false,
		grpcServer}
}

// Init
func (grpcService *GrpcService) Init() {
	log.WithFields(log.Fields{
		"method": grpcService.Name() + ".Init",
	}).Info("initializing...")
}

// Name
func (grpcService *GrpcService) Name() string {
	return "GrpcService"
}

// IsRunning
func (grpcService *GrpcService) IsRunning() bool {
	return grpcService.running
}

// RegisterGrpc
func (grpcService *GrpcService) RegisterGrpc(grpcServer *grpc.Server) {
}

// Go
func (grpcService *GrpcService) Go(waitGroup *sync.WaitGroup) {

	// Create TCP listener/GRPC grpcServer.
	listener, error := net.Listen("tcp", ":"+strconv.Itoa(properties.Properties.GrpcPort))
	if error != nil {
		log.Fatalf("failed to listen: %v", error)
	}

	// Serve.
	reflection.Register(grpcService.grpcServer)
	log.WithFields(log.Fields{
		"method": grpcService.Name() + ".Go",
	}).Info("listening on " + strconv.Itoa(properties.Properties.GrpcPort))
	if error := grpcService.grpcServer.Serve(listener); error != nil {
		log.Fatalf("failed to serve: %v", error)
	}
}
