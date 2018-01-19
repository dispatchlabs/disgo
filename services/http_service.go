package services

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"sync"
	"github.com/dispatchlabs/disgo_node/configurations"
)

type HttpService struct {
	HostIp string
	Port int
	running bool
}

func NewHttpService() *HttpService {

	httpService := HttpService{configurations.Configuration.HttpHostIp, configurations.Configuration.HttpPort, false}

	http.HandleFunc("/", httpService.HandleIndex)

	return &httpService
}

func (httpService *HttpService) Name() string {
	return "HttpService"
}

func (httpService *HttpService) IsRunning() bool {
	return httpService.running
}

func (httpService *HttpService) Go(waitGroup *sync.WaitGroup) {
	httpService.running = true
	listen := httpService.HostIp + ":" + strconv.Itoa(httpService.Port)
	log.WithFields(log.Fields{
		"struct": httpService.Name(),
		"method": "Go",
	}).Info("listening on http://" + listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func (httpService *HttpService) HandleIndex(w http.ResponseWriter, r *http.Request) {
	httpService.httpHeaders(w)
	io.WriteString(w, "hello, world<br/><br/>")
}

func (httpService *HttpService) httpHeaders(responseWriter http.ResponseWriter) {
	responseWriter.Header().Set("Content-Type", "text/html; charset=UTF-8")
}