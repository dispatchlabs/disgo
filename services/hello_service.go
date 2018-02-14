package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dispatchlabs/disgo/party"
	"github.com/dispatchlabs/disgover"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type HelloService struct {
	running bool
}

func NewHelloService() *HelloService {
	return &HelloService{
		running: false,
	}
}

func (HelloService *HelloService) Name() string {
	return "HelloService"
}

func (helloService *HelloService) IsRunning() bool {
	return helloService.running
}

func (helloService *HelloService) Go(waitGroup *sync.WaitGroup) {
	helloService.running = true

	if disgover.GetInstance().ThisContact.Id != "NODE-Seed-001" {
		time.Sleep(time.Second * 5)

		theNode, _ := disgover.GetInstance().Find("NODE-Sample-001", disgover.GetInstance().ThisContact)

		if theNode != nil {
			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", theNode.Endpoint.Host, theNode.Endpoint.Port), grpc.WithInsecure())
			if err != nil {
				log.Fatalf("HelloService -> cannot dial server: %v", err)
			} else {
				log.Info(fmt.Sprintf("HelloService -> connected to %s @ [%s : %d]", theNode.Id, theNode.Endpoint.Host, theNode.Endpoint.Port))
			}

			p := party.NewPartyClient(conn)
			val, _ := p.GetVersion(context.Background(), &party.Empty{})
			fmt.Printf("HelloService -> Remote Node Version %s\n", val)
		}
	}
}
