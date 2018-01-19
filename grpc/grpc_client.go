package grpc

import (
	"github.com/dispatchlabs/disgo/configurations"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	protocolBuffer "github.com/dispatchlabs/disgo/grpc/proto"
	"strconv"
	log "github.com/sirupsen/logrus"
	"time"
)

type GrpcClient struct {
	connection *grpc.ClientConn;
	client protocolBuffer.DisgoGrpcClient
}

func NewGrpcClient(address string) *GrpcClient {

	grpcClient := &GrpcClient{}
	grpc.ConnectionTimeout(time.Second * 10)
	addressString := address + ":" + strconv.Itoa(configurations.Configuration.GrpcPort)
	con, error := grpc.Dial(addressString, grpc.WithInsecure())
	if error != nil {
		log.Fatalf("did not connect: %v", error)
	}

	grpcClient.connection = con
	grpcClient.client = protocolBuffer.NewDisgoGrpcClient(grpcClient.connection)
	log.WithFields(log.Fields{
		"method": "NewGrpcClient",
	}).Info("connected to " + addressString)

	return grpcClient;
}

func (grpcClient *GrpcClient) Send(json string) string {

	response, error := grpcClient.client.Send(context.Background(), &protocolBuffer.GetRequest{Json: json})
	if error != nil {
		log.Fatalf("could not greet: %v", error)
	}

	return response.Json
}


