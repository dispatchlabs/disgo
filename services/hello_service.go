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

	// 	node1, _ := disgover.GetInstance().Find("Bob-Node", disgover.GetInstance().ThisContact)

	// 	if node1 != nil {
	// 		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node1.Endpoint.Host, node1.Endpoint.Port), grpc.WithInsecure())
	// 		if err != nil {
	// 			log.Fatalf("cannot dial server: %v", err)
	// 		}

	// 		p := party.NewPartyClient(conn)
	// 		val, _ := p.GetVersion(context.Background(), &party.Empty{})
	// 		fmt.Printf("Party version %s\n", val)
	// 	}
	// }
}
