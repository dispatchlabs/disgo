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
	connection protocolBuffer.DisgoGrpcClient
}

func NewGrpcClient(address string) *GrpcClient {

	grpcClient := &GrpcClient{}
	grpc.ConnectionTimeout(time.Second * 10)
	addressString := address + ":" + strconv.Itoa(configurations.Configuration.GrpcPort)
	conn, err := grpc.Dial(addressString, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//defer conn.Close()

	grpcClient.connection = protocolBuffer.NewDisgoGrpcClient(conn)
	log.WithFields(log.Fields{
		"method": "NewGrpcClient",
	}).Info("connected to " + addressString)

	return grpcClient;
}

func (grpcClient *GrpcClient) Send(json string) string {

	response, error := grpcClient.connection.Send(context.Background(), &protocolBuffer.GetRequest{Json: json})
	if error != nil {
		log.Fatalf("could not greet: %v", error)
	}

	return response.Json
}
