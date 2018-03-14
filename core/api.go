package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/dispatchlabs/dapos"
	daposCore "github.com/dispatchlabs/dapos/core"

	httpService "github.com/dispatchlabs/commons/services"
	"github.com/dispatchlabs/commons/types"
	"github.com/dispatchlabs/disgover"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Api
type Api struct {
	services []types.IService
	router   *mux.Router
}

// NewApi
func NewApi(services []types.IService) *Api {
	this := Api{services, httpService.GetHttpRouter()}
	this.router.HandleFunc("/v1/ping", this.pingPongHandler).Methods("POST")
	this.router.HandleFunc("/v1/state/{address}", this.retrieveStateHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions/{wallet_address}", this.retrieveTransactionsHandler).Methods("GET")
	this.router.HandleFunc("/v1/transactions", this.createTransactionHandler).Methods("POST")
	return &this

}

func (this *Api) retrieveStateHandler(responseWriter http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	state := dapos.GetDAPoS().GetState(vars["address"])
	if state == nil {
		responseWriter.Write([]byte(`{"status":"STATE_NOT_FOUND"}`))
	}


}

// createTransactionHandler
func (this *Api) createTransactionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, error := ioutil.ReadAll(request.Body)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Api.createTransactionHandler",
		}).Error("unable to read HTTP body of request [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"INTERNAL_SERVER_ERROR"}`, http.StatusInternalServerError)
		return
	}

	// Unmarshal transaction?
	transaction := &daposCore.Transaction{}
	error = json.Unmarshal(body, transaction)
	if error != nil {
		log.WithFields(log.Fields{
			"method": "Api.createTransactionHandler",
		}).Error("JSON parse error [error=" + error.Error() + "]")
		http.Error(responseWriter, `{"status":"JSON_PARSE_ERROR"}`, http.StatusBadRequest)
		return
	}

	if dapos.GetDAPoS().ProcessTxSync(transaction) {
		responseWriter.Write([]byte(`{"status":"OK"}`))
	} else {
		responseWriter.Write([]byte(`{"status":"INVALID_TRANSACTION"}`))
	}
}

func (this *Api) retrieveTransactionsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	//vars := mux.Vars(request)
}

func (this *Api) getService(serviceInterface interface{}) types.IService {
	for _, service := range this.services {
		if reflect.TypeOf(service) == reflect.TypeOf(serviceInterface) {
			return service
		}
	}
	return nil
}

func (this *Api) pingPongHandler(responseWriter http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)

	fmt.Println(string(body))

	responseWriter.Write([]byte(fmt.Sprintf(
		"PONG-From: %s @ %s:%d",
		disgover.GetDisgover().ThisContact.Address,
		disgover.GetDisgover().ThisContact.Endpoint.Host,
		disgover.GetDisgover().ThisContact.Endpoint.Port,
	)))
}
