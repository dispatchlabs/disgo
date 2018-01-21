package grpc

import (
	"github.com/dispatchlabs/disgo/configs"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	protocolBuffer "github.com/dispatchlabs/disgo/grpc/proto"
	"strconv"
	log "github.com/sirupsen/logrus"
	"time"
)

type GrpcClient struct {
	Connection *grpc.ClientConn;
	Client protocolBuffer.DisgoGrpcClient
}

func NewGrpcClient(address string) *GrpcClient {

	grpc.ConnectionTimeout(time.Second * time.Duration(configs.Config.GrpcTimeout))
	addressString := address + ":" + strconv.Itoa(configs.Config.GrpcPort)
	connection, error := grpc.Dial(addressString, grpc.WithInsecure())
	if error != nil {
		log.Fatalf("did not connect: %v", error)
	}

	grpcClient := &GrpcClient{connection, protocolBuffer.NewDisgoGrpcClient(connection)}
	log.WithFields(log.Fields{
		"method": "NewGrpcClient",
	}).Info("connected to " + addressString)
	return grpcClient;
}

func (grpcClient *GrpcClient) Send(json string) string {
	response, error := grpcClient.Client.Send(context.Background(), &protocolBuffer.GetRequest{Json: json})
	if error != nil {
		log.Fatalf("unable to send RPC message: %v", error)
	}
	return response.Json
}

func (grpcClient *GrpcClient) Close(json string) {
	grpcClient.Connection.Close()
}


