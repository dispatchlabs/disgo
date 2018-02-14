package services

import (
	"sync"
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

	// if disgover.GetInstance().ThisContact.Id != "NODE-1" {
	// 	time.Sleep(time.Second * 5)

	// 	theNode, _ := disgover.GetInstance().Find("NODE-Ubuntu", disgover.GetInstance().ThisContact)

	// 	if theNode != nil {
	// 		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "172.18.13.22", theNode.Endpoint.Port), grpc.WithInsecure())
	// 		if err != nil {
	// 			log.Fatalf("HelloService -> cannot dial server: %v", err)
	// 		} else {
	// 			log.Info(fmt.Sprintf("HelloService -> connected to %s @ [%s : %d]", theNode.Id, theNode.Endpoint.Host, theNode.Endpoint.Port))
	// 		}

	// 		p := party.NewPartyClient(conn)
	// 		val, _ := p.GetVersion(context.Background(), &party.Empty{})
	// 		fmt.Printf("HelloService -> Remote Node Version %s\n", val)
	// 	}
	// }
}
