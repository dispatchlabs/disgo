package core

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	commonTypes "github.com/dispatchlabs/disgo_commons/types"
	"github.com/dispatchlabs/disgover"
)

type PingPongService struct {
	running bool
}

func NewPingPongService() *PingPongService {
	return &PingPongService{false}
}

// Name
func (this *PingPongService) Name() string {
	return "PingPongService"
}

// IsRunning
func (this *PingPongService) IsRunning() bool {
	return this.running
}

// Go
func (this *PingPongService) Go(waitGroup *sync.WaitGroup) {
	this.running = true

	go func() {
		time.Sleep(time.Second * 5)

		var contacts = disgover.GetDisgover().GetContactList()

		if len(contacts) > 0 {
			// gen random nr
			s1 := rand.NewSource(time.Now().UnixNano())
			r1 := rand.New(s1)

			// pick a random node
			var randomIndex = r1.Intn(len(contacts))
			var contact *commonTypes.Contact = contacts[randomIndex]

			// send request
			req, _ := http.NewRequest(
				"POST",
				fmt.Sprintf("%s:%d", contact.Endpoint.Host, contact.Endpoint.Port),
				strings.NewReader(fmt.Sprintf(
					"PING-From: %s @ %s:%d",
					disgover.GetDisgover().ThisContact.Address,
					disgover.GetDisgover().ThisContact.Endpoint.Host,
					disgover.GetDisgover().ThisContact.Endpoint.Port,
				)),
			)

			client := &http.Client{}
			client.Do(req)
		}
	}()
}
