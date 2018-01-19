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

	rpcService := RpcService{configurations.Configuration.RpcPort, false}

	return &rpcService
}

func (rpcService *RpcService) Name() string {
	return "RpcService"
}

func (rpcService *RpcService) IsRunning() bool {
	return rpcService.running
}

type server struct{}

func (s *server) Send(ctx context.Context, in *protocolBuffer.GetRequest) (*protocolBuffer.SendResponse, error) {
	return &protocolBuffer.SendResponse{Json: "Hello " + in.Json}, nil
}

func (rpcService *RpcService) Go(waitGroup *sync.WaitGroup) {

	rpcService.running = true
	lis, err := net.Listen("tcp", ":" + strconv.Itoa(rpcService.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	protocolBuffer.RegisterDisgoGrpcServer(s, &server{})
	reflection.Register(s)

	log.WithFields(log.Fields{
		"method": rpcService.Name() + ".Go",
	}).Info("listening on " + strconv.Itoa(rpcService.Port))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
