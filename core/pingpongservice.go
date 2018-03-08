package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/dispatchlabs/disgo_commons/config"
	"github.com/dispatchlabs/disgo_commons/utils"
	"github.com/dispatchlabs/disgover"
	log "github.com/sirupsen/logrus"
)

type PingPongService struct {
	running bool
}

func NewPingPongService() *PingPongService {
	return &PingPongService{false}
}

// IsRunning
func (this *PingPongService) IsRunning() bool {
	return this.running
}

// Go
func (this *PingPongService) Go(waitGroup *sync.WaitGroup) {
	this.running = true

	log.WithFields(log.Fields{
		"method": utils.CallPath(),
	}).Info("running")

	go func() {
		for {
			time.Sleep(time.Minute)

			var contacts = disgover.GetDisgover().GetContactList()

			if len(contacts) > 0 {
				// gen random nr
				s1 := rand.NewSource(time.Now().UnixNano())
				r1 := rand.New(s1)

				// pick a random node
				var randomIndex = r1.Intn(len(contacts))
				var contact *disgover.Contact = contacts[randomIndex]

				var contactUrl = fmt.Sprintf("http://%s:%d/v1/ping", contact.Endpoint.Host, config.Properties.HttpPort) // contact.Endpoint.Port)
				var data = fmt.Sprintf(
					"PING-From: %s @ %s:%d",
					disgover.GetDisgover().ThisContact.Address,
					disgover.GetDisgover().ThisContact.Endpoint.Host,
					disgover.GetDisgover().ThisContact.Endpoint.Port,
				)

				// send request
				req, _ := http.NewRequest(
					"POST",
					contactUrl,
					bytes.NewBuffer([]byte(data)),
				)

				// Send PING
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()

				// Got PONG
				body, _ := ioutil.ReadAll(resp.Body)
				fmt.Println(string(body))
			}
		}
	}()
}
