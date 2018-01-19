package services

import (
	"strconv"
	"sync"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/dispatchlabs/disgo/configurations"
	protocolBuffer "github.com/dispatchlabs/disgo/grpc/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/reflection"
)

type RpcService struct {
	Port int
	running bool
}

func NewRpcService() *RpcService {

	rpcService := RpcService{configurations.Configuration.GrpcPort, false}

	return &rpcService
}

func (rpcService *RpcService) Name() string {
	return "RpcService"
}

func (rpcService *RpcService) IsRunning() bool {
	return rpcService.running
}

type GrpcServer struct{}

func (s *GrpcServer) Send(ctx context.Context, in *protocolBuffer.GetRequest) (*protocolBuffer.SendResponse, error) {
	return &protocolBuffer.SendResponse{Json: "Hello " + in.Json}, nil
}

func (rpcService *RpcService) Go(waitGroup *sync.WaitGroup) {

	rpcService.running = true
	listen, error := net.Listen("tcp", ":" + strconv.Itoa(rpcService.Port))
	if error != nil {
		log.Fatalf("failed to listen: %v", error)
	}

	server := grpc.NewServer()
	protocolBuffer.RegisterDisgoGrpcServer(server, &GrpcServer{})
	reflection.Register(server)
	log.WithFields(log.Fields{
		"method": rpcService.Name() + ".Go",
	}).Info("listening on " + strconv.Itoa(rpcService.Port))
	if error := server.Serve(listen); error != nil {
		log.Fatalf("failed to serve: %v", error)
		rpcService.running = false
	}
}
